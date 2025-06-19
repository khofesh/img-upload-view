import { Link } from "react-router";
import type { Route } from "./+types/home";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Image Upload & View App" },
    { name: "description", content: "Upload and view your images" },
  ];
}

export default function Home() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="max-w-2xl mx-auto px-4 text-center">
        <div className="bg-white rounded-lg shadow-md p-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Image Upload & View
          </h1>
          <p className="text-gray-600 mb-8 text-lg">
            Upload your JPEG images and browse them in gallery
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Link
              to="/upload"
              className="bg-blue-600 text-white px-8 py-4 rounded-lg hover:bg-blue-700 transition-colors shadow-md hover:shadow-lg transform hover:-translate-y-1 transition-transform"
            >
              <div className="text-2xl mb-2">📸</div>
              <div className="font-semibold text-lg mb-1">Upload Images</div>
              <div className="text-sm opacity-90">
                Add new JPEG images (max 10MB)
              </div>
            </Link>

            <Link
              to="/gallery"
              className="bg-green-600 text-white px-8 py-4 rounded-lg hover:bg-green-700 transition-colors shadow-md hover:shadow-lg transform hover:-translate-y-1 transition-transform"
            >
              <div className="text-2xl mb-2">🖼️</div>
              <div className="font-semibold text-lg mb-1">View Gallery</div>
              <div className="text-sm opacity-90">
                Browse all uploaded images
              </div>
            </Link>
          </div>

          <div className="mt-8 pt-6 border-t border-gray-200">
            <p className="text-sm text-gray-500">
              Built with React Router, Go, and PostgreSQL
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
