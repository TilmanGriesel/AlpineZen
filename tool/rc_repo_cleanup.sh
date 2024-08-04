#!/bin/bash

REPO="TilmanGriesel/AlpineZen"

delete_releases() {
  echo "Fetching all releases..."
  releases=$(gh api -X GET "repos/$REPO/releases" --jq '.[].id')

  if [ -z "$releases" ]; then
    echo "No releases found."
  else
    for release_id in $releases; do
      echo "Deleting release ID: $release_id"
      gh api -X DELETE "repos/$REPO/releases/$release_id"
      if [ $? -eq 0 ]; then
        echo "Successfully deleted release ID: $release_id"
      else
        echo "Failed to delete release ID: $release_id"
      fi
    done
  fi
}


delete_tags() {
  echo "Deleting all remote tags..."
  remote_tags=$(git ls-remote --tags origin | awk '{print $2}' | sed 's|refs/tags/||')
  if [ -n "$remote_tags" ]; then
    for tag in $remote_tags; do
      git push origin :refs/tags/$tag
      if [ $? -eq 0 ]; then
        echo "Successfully deleted remote tag: $tag"
      else
        echo "Failed to delete remote tag: $tag"
      fi
    done
  else
    echo "No remote tags found."
  fi

  echo "Fetching all remote tags..."
  git fetch --tags

  echo "Deleting all local tags..."
  git tag -d $(git tag -l)
}

delete_workflow_runs() {
while true; do
  echo "Deleting workflow runs..."
  gh run list --json databaseId -q '.[].databaseId' | while read -r id; do
    gh api -X DELETE "repos/$(gh repo view --json nameWithOwner -q .nameWithOwner)/actions/runs/$id"
  done
done
}

export PAGER='cat'

confirm_execution() {
  read -p "Are you sure you want to proceed? Type YES to continue: " confirmation
  if [ "$confirmation" != "YES" ]; then
    echo "Operation aborted."
    exit 1
  fi
}
confirm_execution

# Run cleanpp
delete_releases
delete_tags
delete_workflow_runs
