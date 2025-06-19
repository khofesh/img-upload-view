import { useEffect, useState } from "react";
import { Link } from "react-router";
import type { Route } from "./+types/gallery";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Image Gallery" },
    { name: "description", content: "Browse uploaded images" },
  ];
}

interface Image {
  id: number;
  filename: string;
  original_filename: string;
  url: string;
  file_size: number;
  content_type: string;
  upload_timestamp: string;
}

interface GalleryResponse {
  images: Image[];
  metadata: {
    total_count: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

export default function Gallery() {
  const [images, setImages] = useState<Image[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedImage, setSelectedImage] = useState<Image | null>(null);

  useEffect(() => {
    fetchImages();
  }, []);

  const fetchImages = async () => {
    try {
      setLoading(true);
      setError(null);

      const apiUrl = import.meta.env.VITE_API_URL || "/api";
      console.log(apiUrl);
      const response = await fetch(`${apiUrl}/images?limit=20&offset=0`);

      if (!response.ok) {
        throw new Error("Failed to fetch images");
      }

      const data: GalleryResponse = await response.json();
      setImages(data.images);
    } catch (err) {
      console.error("Error fetching images:", err);
      setError(err instanceof Error ? err.message : "Failed to load images");
    } finally {
      setLoading(false);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const openModal = (image: Image) => {
    setSelectedImage(image);
  };

  const closeModal = () => {
    setSelectedImage(null);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading images...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="bg-red-50 border border-red-200 rounded-md p-6 max-w-md">
            <p className="text-red-800 font-semibold mb-2">
              Error Loading Images
            </p>
            <p className="text-red-600 text-sm">{error}</p>
            <button
              onClick={fetchImages}
              className="mt-4 bg-red-600 text-white px-4 py-2 rounded-md hover:bg-red-700 transition-colors"
            >
              Try Again
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4">
        {/* header */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Image Gallery
          </h1>
          <p className="text-gray-600">
            {images.length} image{images.length !== 1 ? "s" : ""} uploaded
          </p>
        </div>

        {/* nav */}
        <div className="mb-8">
          <div className="flex justify-between items-center">
            <Link
              to="/"
              className="text-blue-600 hover:text-blue-800 font-medium"
            >
              ← Back to Home
            </Link>
            <Link
              to="/upload"
              className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
            >
              Upload New Image
            </Link>
          </div>
        </div>

        {/* empty images */}
        {images.length === 0 ? (
          <div className="text-center py-12">
            <div className="bg-white rounded-lg shadow-md p-8 max-w-md mx-auto">
              <p className="text-gray-600 mb-4">No images uploaded yet</p>
              <Link
                to="/upload"
                className="bg-blue-600 text-white px-6 py-3 rounded-md hover:bg-blue-700 transition-colors inline-block"
              >
                Upload Your First Image
              </Link>
            </div>
          </div>
        ) : (
          /* image grid */
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {images.map((image) => (
              <div
                key={image.id}
                className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow cursor-pointer"
                onClick={() => openModal(image)}
              >
                <div className="aspect-square bg-gray-100">
                  <img
                    src={image.url}
                    alt={image.original_filename}
                    className="w-full h-full object-cover"
                    loading="lazy"
                  />
                </div>
                <div className="p-4">
                  <h3
                    className="font-semibold text-gray-900 truncate"
                    title={image.original_filename}
                  >
                    {image.original_filename}
                  </h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {formatFileSize(image.file_size)}
                  </p>
                  <p className="text-xs text-gray-400 mt-1">
                    {formatDate(image.upload_timestamp)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* modal */}
        {selectedImage && (
          <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg max-w-4xl max-h-full overflow-auto">
              <div className="p-4 border-b border-gray-200 flex justify-between items-center">
                <h2 className="text-xl font-semibold text-gray-900">
                  {selectedImage.original_filename}
                </h2>
                <button
                  onClick={closeModal}
                  className="text-gray-400 hover:text-gray-600 text-2xl"
                >
                  ×
                </button>
              </div>
              <div className="p-4">
                <img
                  src={selectedImage.url}
                  alt={selectedImage.original_filename}
                  className="max-w-full h-auto mx-auto"
                />
                <div className="mt-4 text-sm text-gray-600 space-y-1">
                  <p>
                    <strong>File Size:</strong>{" "}
                    {formatFileSize(selectedImage.file_size)}
                  </p>
                  <p>
                    <strong>Type:</strong> {selectedImage.content_type}
                  </p>
                  <p>
                    <strong>Uploaded:</strong>{" "}
                    {formatDate(selectedImage.upload_timestamp)}
                  </p>
                  <p>
                    <strong>Filename:</strong> {selectedImage.filename}
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
