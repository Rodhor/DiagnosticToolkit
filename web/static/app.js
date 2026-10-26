document.addEventListener("htmx:afterSettle", (event) => {
	const target = event.target;
	if (!(target instanceof HTMLElement) || !target.matches("[data-results-poll]")) return;

	const rows = target.querySelectorAll("tbody tr");
	const latest = rows.length ? rows[rows.length - 1].textContent.replace(/\s+/g, " ").trim() : "";
	const message = latest ? `${rows.length} observations. Latest: ${latest}` : "No active observations.";
	const announcement = document.getElementById("results-announcement");
	if (announcement && announcement.textContent !== message) announcement.textContent = message;
});
