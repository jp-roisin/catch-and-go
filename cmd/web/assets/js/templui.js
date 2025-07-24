document.addEventListener("DOMContentLoaded", () => {
  // Re-initialize templUI components after HTMX swaps
  document.body.addEventListener("htmx:afterSwap", (e) => {
    if (window.templUI) {
      Object.values(window.templUI).forEach(comp => {
        comp.init?.(e.detail.elt);
      });
    }
  });

  // Re-initialize components after out-of-band swaps
  document.body.addEventListener("htmx:oobAfterSwap", (e) => {
    if (window.templUI) {
      Object.values(window.templUI).forEach(comp => {
        comp.init?.(e.detail.target);
      });
    }
  });

  // Cleanup before swap (for components with event listeners)
  document.body.addEventListener("htmx:beforeSwap", (e) => {
    if (window.templUI) {
      Object.values(window.templUI).forEach(comp => {
        comp.cleanup?.(e.detail.target);
      });
    }
  });
});
