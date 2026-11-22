export function formatWhen(iso) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      year: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}

export function formatCoord(value) {
  return Number(value).toFixed(5)
}

export function osmUrl(lat, lng) {
  const latitude = formatCoord(lat)
  const longitude = formatCoord(lng)
  return `https://www.openstreetmap.org/?mlat=${latitude}&mlon=${longitude}#map=15/${latitude}/${longitude}`
}

export function formatDelta(current, previous) {
  if (previous == null || current === previous) return null
  const diff = current - previous
  return {
    text: `${diff > 0 ? '+' : '−'}${Math.abs(diff)}`,
    kind: diff > 0 ? 'up' : 'down',
  }
}
