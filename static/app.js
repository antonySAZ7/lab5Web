async function nextEpisode(id) {
    await fetch(`/update?id=${id}`, { method: "POST" })
    location.reload()
}

async function prevEpisode(id) {
    await fetch(`/decrement?id=${id}`, { method: "POST" })
    location.reload()
}