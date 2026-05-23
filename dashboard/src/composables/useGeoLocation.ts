import { ref, onMounted } from 'vue'

export interface GeoInfo {
  lat: number; lng: number
  country: string; region: string; city: string
}

export function useGeoLocation() {
  const location = ref<GeoInfo | null>(null)

  onMounted(async () => {
    // 用 IP 地理定位获取详细信息
    try {
      const resp = await fetch('http://ip-api.com/json/?fields=lat,lon,country,regionName,city', { signal: AbortSignal.timeout(5000) })
      const data = await resp.json()
      if (data.lat && data.lon) {
        location.value = { lat: data.lat, lng: data.lon, country: data.country || '', region: data.regionName || '', city: data.city || '' }
        return
      }
    } catch { /* 网络不可达 */ }

    // 兜底：浏览器 Geolocation API
    if ('geolocation' in navigator) {
      try {
        const pos = await new Promise<GeolocationPosition>((resolve, reject) => {
          navigator.geolocation.getCurrentPosition(resolve, reject, { timeout: 5000, maximumAge: 600000 })
        })
        location.value = { lat: pos.coords.latitude, lng: pos.coords.longitude, country: '', region: '', city: '' }
      } catch { /* 用户拒绝 */ }
    }
  })

  return { location }
}
