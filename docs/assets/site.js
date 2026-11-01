(() => {
  "use strict";
  const menuToggle = document.querySelector(".menu-toggle");
  const primaryNav = document.getElementById("primary-nav");
  function closeMenu() {
    if (!menuToggle || !primaryNav) return;
    menuToggle.setAttribute("aria-expanded", "false");
    primaryNav.classList.remove("is-open");
  }
  if (menuToggle && primaryNav) {
    menuToggle.addEventListener("click", () => {
      const show = menuToggle.getAttribute("aria-expanded") !== "true";
      menuToggle.setAttribute("aria-expanded", String(show));
      primaryNav.classList.toggle("is-open", show);
    });
    primaryNav.addEventListener("click", event => {
      if (event.target.closest("a")) closeMenu();
    });
    document.addEventListener("keydown", event => {
      if (event.key === "Escape" && primaryNav.classList.contains("is-open")) {
        closeMenu();
        menuToggle.focus();
      }
    });
  }

  const search = document.querySelector("[data-doc-search]");
  if (search) {
    const cards = [...document.querySelectorAll("[data-doc-entry]")];
    const sections = [...document.querySelectorAll("[data-guide-section]")];
    const status = document.querySelector("[data-search-status]");
    const noResults = document.querySelector("[data-no-results]");
    function filter() {
      const query = search.value.toLocaleLowerCase().trim();
      let count = 0;
      for (const card of cards) {
        const haystack = (card.textContent + " " + (card.dataset.searchTags || "")).toLocaleLowerCase();
        const visible = !query || haystack.includes(query);
        card.hidden = !visible;
        if (visible) count++;
      }
      for (const section of sections) {
        section.hidden = ![...section.querySelectorAll("[data-doc-entry]")].some(card => !card.hidden);
      }
      if (status) status.textContent = query
        ? count + (count === 1 ? " guide found" : " guides found")
        : "Browse all " + cards.length + " guides";
      if (noResults) noResults.hidden = count !== 0;
    }
    search.addEventListener("input", filter);
    document.addEventListener("keydown", event => {
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        search.focus();
      }
    });
    filter();
  }
  // Progressive enhancement for code examples; leave code readable without JS.
  for (const block of document.querySelectorAll(".prose pre")) {
    const code = block.querySelector("code");
    if (!code) continue;
    const button = document.createElement("button");
    button.type = "button";
    button.className = "copy-code";
    button.textContent = "Copy";
    button.setAttribute("aria-label", "Copy code example");
    button.addEventListener("click", async () => {
      try {
        await navigator.clipboard.writeText(code.innerText || code.textContent || "");
        button.textContent = "Copied";
        button.setAttribute("aria-label", "Code copied");
        button.addEventListener("blur", () => {
          button.textContent = "Copy";
          button.setAttribute("aria-label", "Copy code example");
        }, {once:true});
      } catch {
        button.textContent = "Select text to copy";
      }
    });
    block.appendChild(button);
  }
  for (const heading of document.querySelectorAll(".prose h2[id], .prose h3[id]")) {
    const link = document.createElement("a");
    link.className = "permalink-icon";
    link.href = "#" + encodeURIComponent(heading.id);
    link.textContent = "#";
    link.setAttribute("aria-label", "Link to " + heading.textContent.trim());
    heading.appendChild(link);
  }
})();
