(() => {
  const button = document.querySelector(".menu-toggle");
  const menu = document.getElementById("primary-nav");
  if (button && menu) {
    button.addEventListener("click", () => {
      const open = button.getAttribute("aria-expanded") !== "true";
      button.setAttribute("aria-expanded", String(open));
      menu.classList.toggle("is-open", open);
    });
    menu.addEventListener("click", (event) => {
      if (event.target.closest("a")) {
        menu.classList.remove("is-open");
        button.setAttribute("aria-expanded", "false");
      }
    });
    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape" && menu.classList.contains("is-open")) {
        menu.classList.remove("is-open");
        button.setAttribute("aria-expanded", "false");
        button.focus();
      }
    });
  }
  const search = document.querySelector("[data-doc-search]");
  if (search) {
    const cards = [...document.querySelectorAll("[data-doc-entry]")];
    const status = document.querySelector("[data-search-status]");
    const run = () => {
      const query = search.value.toLowerCase().trim();
      let matches = 0;
      cards.forEach(card => {
        const visible = !query || card.textContent.toLowerCase().includes(query);
        card.hidden = !visible;
        if (visible) matches++;
      });
      if (status) status.textContent = query ? matches + " guides found" : "Browse all guides";
    };
    search.addEventListener("input", run);
    document.addEventListener("keydown", event => {
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        search.focus();
      }
    });
  }
})();
