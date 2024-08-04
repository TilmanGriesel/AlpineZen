# SPDX-FileCopyrightText: 2024 Tilman Griesel
#
# SPDX-License-Identifier: GPL-3.0-or-later

import os
import subprocess
import sys

project = "AlpineZen"
organization = "Tilman Griesel"
year = "2024"
license_id = "GPL-3.0-or-later"
license_file_url = f"https://raw.githubusercontent.com/licenses/license-templates/master/templates/gpl3-header.txt"

def create_virtualenv():
    venv_dir = os.path.join(os.getcwd(), '.venv')
    if not os.path.exists(venv_dir):
        subprocess.run([sys.executable, '-m', 'venv', venv_dir])
        print(f"Virtual environment created at {venv_dir}")
    else:
        print(f"Virtual environment already exists at {venv_dir}")
    return venv_dir

def install_reuse(venv_dir):
    pip_path = os.path.join(venv_dir, 'bin', 'pip')
    subprocess.run([pip_path, 'install', 'reuse'])
    print("REUSE tool installed in the virtual environment")


def add_license_headers(venv_dir, include_dirs):
    reuse_path = os.path.join(venv_dir, 'bin', 'reuse')
    project_dir = os.getcwd()

    whitelist_extensions = {'.yaml', '.yml', '.go', '.py', '.sh', '.bat', '.mts', '.ts', '.svg', '.png', '.icns'}

    for directory in include_dirs:
        for root, dirs, files in os.walk(directory):
            for file in files:
                file_path = os.path.join(root, file)

                if os.path.splitext(file)[1].lower() not in whitelist_extensions:
                    continue

                subprocess.run([
                    reuse_path, 'annotate',
                    '--license', license_id,
                    '--copyright', f"{organization}",
                    '--skip-unrecognised',
                    file_path
                ])

    print("License headers added to all relevant files")

def download_licenses(venv_dir, project_dir):
    reuse_path = os.path.join(venv_dir, 'bin', 'reuse')
    subprocess.run([reuse_path, 'download', '--all'])
    print("Download licenses completed")

def check_compliance(venv_dir, project_dir):
    reuse_path = os.path.join(venv_dir, 'bin', 'reuse')
    subprocess.run([reuse_path, '--root', project_dir, 'lint'])
    print("Compliance check completed")

if __name__ == "__main__":
    project_dir = os.getcwd()
    license_dir = os.path.join(project_dir, 'LICENSES')

    include_dirs = [
        os.path.join(project_dir, 'cmd/cli'),
        os.path.join(project_dir, 'pkg'),
        os.path.join(project_dir, 'assets'),
        os.path.join(project_dir, 'docs'),
        os.path.join(project_dir, '.github'),
    ]

    venv_dir = create_virtualenv()
    install_reuse(venv_dir)
    download_licenses(venv_dir, project_dir)
    add_license_headers(venv_dir, include_dirs)
    check_compliance(venv_dir, project_dir)
