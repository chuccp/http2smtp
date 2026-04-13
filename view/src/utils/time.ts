/**
 * 时间格式化工具
 */

/**
 * 格式化时间为 "2022-11-21 12:22:26" 格式
 * @param time - 时间
 * @param format - 格式，默认 "YYYY-MM-DD HH:mm:ss"
 * @returns 格式化后的时间字符串
 */
export function formatTime(time: string | Date | null | undefined, format: string = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!time) return ''

  let date: Date
  if (typeof time === 'string') {
    // 处理 ISO 格式或其他格式
    date = new Date(time)
  } else if (time instanceof Date) {
    date = time
  } else {
    return ''
  }

  if (isNaN(date.getTime())) return ''

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')

  return format
    .replace('YYYY', String(year))
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

/**
 * 格式化为日期 "2022-11-21"
 */
export function formatDate(time: string | Date | null | undefined): string {
  return formatTime(time, 'YYYY-MM-DD')
}

export default {
  formatTime,
  formatDate
}
