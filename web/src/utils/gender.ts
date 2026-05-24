export function genderCodeToLabel(code: number) {
  return code === 1 ? '男' : '女'
}

export function genderLabelToCode(label: string) {
  return label === '男' ? 1 : 0
}
