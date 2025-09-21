import { ReactNode } from 'react';

export default function BooksPageTemplate({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-slate-50">
      <div className="mx-auto max-w-5xl p-6 space-y-6">
        <header className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Books</h1>
          <a href="/" className="text-slate-600 underline">Home</a>
        </header>
        {children}
      </div>
    </div>
  );
}