import { Book } from '@/types/book';

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || '';


async function http<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  });
  if (!res.ok) throw new Error(await res.text());
  if (res.status === 204) return undefined as unknown as T;
  return res.json();
}

export const api = {
  listBooks: () => http<Book[]>('/books/'),
  getBook: (id: string | number) => http<Book>(`/books/${id}`),
  createBook: (payload: Partial<Book>) =>
    http<Book>('/books/', { method: 'POST', body: JSON.stringify(payload) }),
  updateBook: (id: string | number, payload: Partial<Book>) =>
    http<Book>(`/books/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  deleteBook: (id: string | number) => http<void>(`/books/${id}`, { method: 'DELETE' }),
};