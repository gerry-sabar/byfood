import { z } from "zod";

export const bookSchema = z.object({
  title: z
    .string()
    .min(1, "Title is required")
    .max(120, "Title must be ≤ 120 characters"),

  author: z
    .string()
    .min(1, "Author is required")
    .max(80, "Author must be ≤ 80 characters"),

  isbn: z
    .string()
    .min(1, "ISBN is required")
    .regex(/^(97(8|9))?\d{9}(\d|X)$/, "Invalid ISBN"), // adjust if you already have custom validator

  publication_year: z
    .number({
      required_error: "Publication year is required",
      invalid_type_error: "Publication year must be a number",
    })
    .int("Publication year must be an integer")
    .gte(1000, "Publication year must be 4 digits")
    .lte(9999, "Publication year must be 4 digits"),

  price: z
    .number({
      required_error: "Price is required",
      invalid_type_error: "Price must be a number",
    })
    .nonnegative("Price must be ≥ 0")
    .max(1_000_000, "Price too large")
    .refine((val) => Number.isFinite(val) && Math.round(val * 100) === val * 100, {
      message: "Price must have at most 2 decimal places",
    }),
});

// TypeScript type for forms
export type BookFormValues = z.infer<typeof bookSchema>;
