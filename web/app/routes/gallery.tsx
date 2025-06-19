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
  const [deleteConfirm, setDeleteConfirm] = useState<Image | null>(null);
  const [deleting, setDeleting] = useState<number | null>(null);

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

  const handleDelete = async (image: Image) => {
    try {
      setDeleting(image.id);
      const apiUrl = import.meta.env.VITE_API_URL || "/api";

      const response = await fetch(`${apiUrl}/image/${image.id}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to delete image");
      }

      // remove from local state
      setImages((prevImages) =>
        prevImages.filter((img) => img.id !== image.id)
      );
      setDeleteConfirm(null);

      // close modal if being shown
      if (selectedImage?.id === image.id) {
        setSelectedImage(null);
      }
    } catch (err) {
      console.error("Error deleting image:", err);
      setError(err instanceof Error ? err.message : "Failed to delete image");
    } finally {
      setDeleting(null);
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

  const openDeleteConfirm = (image: Image, event: React.MouseEvent) => {
    event.stopPropagation();
    setDeleteConfirm(image);
  };

  const closeDeleteConfirm = () => {
    setDeleteConfirm(null);
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
              onClick={() => {
                setError(null);
                fetchImages();
              }}
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
                className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow cursor-pointer relative group"
                onClick={() => openModal(image)}
              >
                <div className="aspect-square bg-gray-100 relative">
                  <img
                    src={image.url}
                    alt={image.original_filename}
                    className="w-full h-full object-cover"
                    loading="lazy"
                  />
                  {/* delete button */}
                  <button
                    onClick={(e) => openDeleteConfirm(image, e)}
                    disabled={deleting === image.id}
                    className="absolute top-2 right-2 bg-red-600 text-white p-2 rounded-full opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-700 disabled:opacity-50"
                    title="Delete image"
                  >
                    {deleting === image.id ? (
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                    ) : (
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    )}
                  </button>
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
                <div className="flex items-center space-x-2">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      openDeleteConfirm(selectedImage, e);
                    }}
                    className="text-red-600 hover:text-red-800 p-2 rounded-md hover:bg-red-50"
                    title="Delete image"
                  >
                    <svg
                      className="w-5 h-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </button>
                  <button
                    onClick={closeModal}
                    className="text-gray-400 hover:text-gray-600 text-2xl"
                  >
                    ×
                  </button>
                </div>
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

        {/*delete confirmation */}
        {deleteConfirm && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg max-w-md w-full">
              <div className="p-6">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mr-4">
                    <svg
                      className="w-6 h-6 text-red-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 18.5c-.77.833.192 2.5 1.732 2.5z"
                      />
                    </svg>
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">
                      Delete Image
                    </h3>
                    <p className="text-sm text-gray-500">
                      This action cannot be undone
                    </p>
                  </div>
                </div>

                <p className="text-gray-700 mb-6">
                  Are you sure you want to delete "
                  {deleteConfirm.original_filename}"?
                </p>

                <div className="flex justify-end space-x-3">
                  <button
                    onClick={closeDeleteConfirm}
                    disabled={deleting === deleteConfirm.id}
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors disabled:opacity-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => handleDelete(deleteConfirm)}
                    disabled={deleting === deleteConfirm.id}
                    className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors disabled:opacity-50 flex items-center"
                  >
                    {deleting === deleteConfirm.id ? (
                      <>
                        <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                        Deleting...
                      </>
                    ) : (
                      "Delete"
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
