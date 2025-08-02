/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        system: [
          "-apple-system",
          "BlinkMacSystemFont",
          '"SF Pro Display"',
          '"SF Pro Text"',
          "system-ui",
          "sans-serif",
        ],
      },
      colors: {
        // macOS Light Mode
        "macos-bg-primary": "rgba(255, 255, 255, 0.8)",
        "macos-bg-secondary": "rgba(246, 246, 246, 0.8)",
        "macos-bg-tertiary": "rgba(242, 242, 247, 0.8)",
        "macos-text-primary": "rgba(0, 0, 0, 0.85)",
        "macos-text-secondary": "rgba(0, 0, 0, 0.6)",
        "macos-text-tertiary": "rgba(0, 0, 0, 0.4)",
        "macos-accent-blue": "#007AFF",
        "macos-accent-red": "#FF3B30",
        "macos-accent-green": "#34C759",
        "macos-border": "rgba(0, 0, 0, 0.1)",

        // macOS Dark Mode
        "macos-dark-bg-primary": "rgba(28, 28, 30, 0.8)",
        "macos-dark-bg-secondary": "rgba(44, 44, 46, 0.8)",
        "macos-dark-bg-tertiary": "rgba(58, 58, 60, 0.8)",
        "macos-dark-text-primary": "rgba(255, 255, 255, 0.85)",
        "macos-dark-text-secondary": "rgba(255, 255, 255, 0.6)",
        "macos-dark-text-tertiary": "rgba(255, 255, 255, 0.4)",
        "macos-dark-accent-blue": "#0A84FF",
        "macos-dark-accent-red": "#FF453A",
        "macos-dark-accent-green": "#32D74B",
        "macos-dark-border": "rgba(255, 255, 255, 0.1)",
      },
      spacing: {
        4.5: "1.125rem", // 18px
        15: "3.75rem", // 60px
      },
      borderRadius: {
        macos: "8px",
        "macos-button": "6px",
        "macos-input": "4px",
      },
      boxShadow: {
        macos: "0 4px 20px rgba(0, 0, 0, 0.1)",
        "macos-dark": "0 4px 20px rgba(0, 0, 0, 0.3)",
      },
      backdropBlur: {
        macos: "20px",
      },
      animation: {
        "slide-in": "slideIn 0.2s cubic-bezier(0.16, 1, 0.3, 1)",
        "scale-in": "scaleIn 0.15s ease-out",
      },
      keyframes: {
        slideIn: {
          "0%": { opacity: "0", transform: "translateY(-10px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        scaleIn: {
          "0%": { opacity: "0", transform: "scale(0.95)" },
          "100%": { opacity: "1", transform: "scale(1)" },
        },
      },
    },
  },
  plugins: [require("@tailwindcss/forms")],
  darkMode: "media", // Responds to system preference
};
