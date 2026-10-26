(function () {
	const root = document.documentElement;
	const storageKey = "diagnostic-toolkit-theme";
	const toggle = document.getElementById("theme-toggle");

	function preferredTheme() {
		const stored = localStorage.getItem(storageKey);
		if (stored === "dark" || stored === "light") return stored;
		return "dark";
	}

	function applyTheme(theme) {
		root.classList.toggle("dark", theme === "dark");
		if (toggle) toggle.setAttribute("aria-pressed", theme === "dark" ? "true" : "false");
	}

	applyTheme(preferredTheme());

	if (toggle) {
		toggle.addEventListener("click", () => {
			const next = root.classList.contains("dark") ? "light" : "dark";
			localStorage.setItem(storageKey, next);
			applyTheme(next);
		});
	}

	document.addEventListener("htmx:afterSettle", (event) => {
		const target = event.target;
		if (!(target instanceof HTMLElement) || !target.matches("[data-results-poll]")) return;
		const rows = target.querySelectorAll("tbody tr");
		const latest = rows.length ? rows[rows.length - 1].textContent.replace(/\s+/g, " ").trim() : "";
		const message = latest ? `${rows.length} observations. Latest: ${latest}` : "No active observations.";
		const announcement = document.getElementById("results-announcement");
		if (announcement && announcement.textContent !== message) announcement.textContent = message;
	});
})();
