<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'
import L from 'leaflet'
import { useAdminStore } from '../stores/admin'
import { escapeHtml } from '../utils/escapeHtml'
import { formatCoord, formatWhen, osmUrl } from '../utils/format'
import { reverseGeocode } from '../utils/geocode'

const store = useAdminStore()
const mapEl = ref(null)
const hasLocated = ref(false)

let map = null
let layer = null
let markerCount = -1
let invalidateTimer = null

function popupHtml(user) {
  const name = escapeHtml(user.name || '—')
  const email = escapeHtml(user.email || '—')
  const when = escapeHtml(formatWhen(user.lastLocationAt))
  const status = user.isActive
    ? '<span class="badge badge-ok">ativo</span>'
    : '<span class="badge badge-off">suspenso</span>'
  const url = osmUrl(user.lastLat, user.lastLng)

  return (
    `<b>${name}</b>` +
    `${email}<br>` +
    `${status} · ${when}<br>` +
    '<span class="popup-addr" data-pending="1">endereço…</span><br>' +
    `<a href="${url}" target="_blank" rel="noopener">abrir no OpenStreetMap</a>`
  )
}

// Preenche a linha de endereço do popup assim que o Nominatim responde.
// Só toca no elemento se ele ainda estiver no DOM (o refresh de 10s recria
// os marcadores e os popups).
function fillPopupAddress(marker, user) {
  marker.on('popupopen', () => {
    const popupEl = marker.getPopup()?.getElement()
    if (!popupEl) return
    const addrEl = popupEl.querySelector('.popup-addr')
    if (!addrEl || !addrEl.dataset.pending) return
    reverseGeocode(user.lastLat, user.lastLng).then((addr) => {
      if (!addrEl.isConnected) return
      addrEl.textContent = addr || 'endereço indisponível'
      delete addrEl.dataset.pending
    })
  })
}

function initMap() {
  if (map || !mapEl.value) return
  map = L.map(mapEl.value, { scrollWheelZoom: true }).setView([-23.5505, -46.6333], 4)
  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
  }).addTo(map)
  layer = L.layerGroup().addTo(map)
  L.Icon.Default.mergeOptions({
    iconUrl: 'vendor/leaflet/images/marker-icon.png',
    iconRetinaUrl: 'vendor/leaflet/images/marker-icon-2x.png',
    shadowUrl: 'vendor/leaflet/images/marker-shadow.png',
  })
}

function renderMarkers() {
  if (!map || !layer) return

  const located = store.users.filter((user) => user.lastLat != null && user.lastLng != null)
  hasLocated.value = located.length > 0

  layer.clearLayers()
  const bounds = []
  located.forEach((user) => {
    const className = user.isActive ? 'map-marker-active' : 'map-marker-inactive'
    const icon = L.divIcon({
      className,
      iconSize: [26, 26],
      iconAnchor: [13, 26],
      popupAnchor: [0, -24],
    })
    const marker = L.marker([user.lastLat, user.lastLng], { icon }).bindPopup(popupHtml(user))
    fillPopupAddress(marker, user)
    layer.addLayer(marker)
    bounds.push([user.lastLat, user.lastLng])
  })

  if (bounds.length && bounds.length !== markerCount) {
    map.fitBounds(bounds, { padding: [40, 40], maxZoom: 13 })
  }
  markerCount = bounds.length

  if (invalidateTimer) clearTimeout(invalidateTimer)
  invalidateTimer = setTimeout(() => map.invalidateSize(), 100)
}

function invalidate() {
  if (map) map.invalidateSize()
}

onMounted(() => {
  initMap()
  renderMarkers()
})

onUnmounted(() => {
  if (invalidateTimer) clearTimeout(invalidateTimer)
  if (map) map.remove()
})

watch(() => store.users, renderMarkers)

defineExpose({ invalidate })
</script>

<template>
  <div class="map-section">
    <h2 class="section-title">Última posição dos usuários</h2>
    <div id="map-wrap">
      <div id="map" ref="mapEl"></div>
      <p id="map-empty" :hidden="hasLocated">Nenhum usuário com posição conhecida ainda.</p>
    </div>
    <p class="hint">
      Marcadores da última posição conhecida (sem histórico). Âmbar = usuário suspenso. Dados:
      OpenStreetMap.
    </p>
  </div>
</template>
