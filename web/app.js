const result = document.querySelector('#result');
async function load() { const response = await fetch('/api/passes'); result.textContent = JSON.stringify(await response.json(), null, 2); }
document.querySelector('#reload').addEventListener('click', load);
document.querySelector('#pass-form').addEventListener('submit', async (event) => { event.preventDefault(); const data = Object.fromEntries(new FormData(event.target)); data.minutes = Number(data.minutes); const response = await fetch('/api/passes', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(data) }); result.textContent = JSON.stringify(await response.json(), null, 2); });
load();
