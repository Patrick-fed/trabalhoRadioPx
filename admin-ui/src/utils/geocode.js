// Reverse geocoding via Nominatim/OSM para o popup do mapa (DEC-020).
// Lazy: só é chamado quando um popup abre. Cache em memória por coordenada
// (4 casas ≈ 11m); falhas resolvem null e também são cacheadas, para um popup
// quebrado não tentar de novo em loop. Política de uso do Nominatim: no
// máximo 1 req/s — fila no nível do módulo, compartilhada por todos os popups.

const cache = new Map() // "lat,lng" (4dp) -> Promise<string|null>
let lastCall = 0

function geoKey(lat, lng) {
  return Number(lat).toFixed(4) + ',' + Number(lng).toFixed(4)
}

export function reverseGeocode(lat, lng) {
  const key = geoKey(lat, lng)
  if (cache.has(key)) return cache.get(key)

  const p = (async () => {
    const wait = lastCall + 1000 - Date.now()
    if (wait > 0) {
      await new Promise((resolve) => setTimeout(resolve, wait))
    }
    lastCall = Date.now()
    try {
      const url =
        'https://nominatim.openstreetmap.org/reverse?format=jsonv2' +
        `&lat=${encodeURIComponent(lat)}` +
        `&lon=${encodeURIComponent(lng)}` +
        '&accept-language=pt-BR&zoom=18'
      const res = await fetch(url, { headers: { Accept: 'application/json' } })
      if (!res.ok) return null
      const data = await res.json()
      return data && data.display_name ? data.display_name : null
    } catch {
      return null
    }
  })()

  cache.set(key, p)
  return p
}
