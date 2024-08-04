import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  // logo: "assets/brand/alpinezen_logo_sun.svg",
  title: "AlpineZen",
  titleTemplate: ":title",
  description:
    "Ascend your digital workspace with AlpineZen, a free, user-friendly tool carefully crafted to make your workday more enjoyable. Whether you're tackling tasks or enjoying some downtime, AlpineZen enhances your screen with dynamic wallpapers that shift and evolve in harmony with the day’s natural rhythms. Create an inspiring and calming workspace that keeps you motivated and centered throughout the day.",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    // nav: [
    //   { text: "Home", link: "/" },
    //   { text: "Examples", link: "/markdown-examples" },
    // ],
    // sidebar: [
    //   {
    //     text: "Examples",
    //     items: [
    //       { text: "Markdown Examples", link: "/markdown-examples" },
    //       { text: "Runtime API Examples", link: "/api-examples" },
    //     ],
    //   },
    // ],

    socialLinks: [
      { icon: "github", link: "https://github.com/TilmanGriesel/AlpineZen" },
    ],
  },
});
