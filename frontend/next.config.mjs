/** @type {import('next').NextConfig} */
const API_BASE = process.env.BACKEND_URL ?? "http://localhost:8080";

const nextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${API_BASE}/:path*`, 
      },
    ];
  },
};

export default nextConfig;
