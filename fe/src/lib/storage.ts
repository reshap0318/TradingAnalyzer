const KEY_TOKEN: string = 'auth_token'

export function getToken(): string {
  return localStorage.getItem(KEY_TOKEN) as string
}

export function setToken(token: string): void {
  window.localStorage.setItem(KEY_TOKEN, token)
}

export function destroySession(): void {
  window.localStorage.clear()
}
