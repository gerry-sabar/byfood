export type Book = {
  id: number;
  title: string;
  author: string;
  isbn: string;
  price: number;
  publication_year: number;
  created_at?: string;
  updated_at?: string;
};