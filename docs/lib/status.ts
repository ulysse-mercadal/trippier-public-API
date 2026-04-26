export function statusClass(s: number): 'ok' | 'err' | '' {
  if (s >= 200 && s < 300) return 'ok';
  if (s >= 400) return 'err';
  return '';
}
