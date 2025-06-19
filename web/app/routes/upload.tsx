import { useState } from "react";
import { Link } from "react-router";
import type { Route } from "./+types/upload";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Upload Image" },
    { name: "description", content: "Upload your images here" },
  ];
}

export default function Upload() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];

    if (!file) {
      setSelectedFile(null);
      setPreviewUrl(null);
      return;
    }

    // check file type
    if (file.type !== "image/jpeg" && file.type !== "image/jpg") {
      setUploadError("Only JPEG images are allowed");
      setSelectedFile(null);
      setPreviewUrl(null);
      return;
    }

    // check file size (10MB)
    if (file.size > 10 * 1024 * 1024) {
      setUploadError("File size must be less than 10MB");
      setSelectedFile(null);
      setPreviewUrl(null);
      return;
    }

    setSelectedFile(file);
    setUploadError(null);
    setUploadSuccess(false);

    const url = URL.createObjectURL(file);
    setPreviewUrl(url);
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setUploading(true);
    setUploadError(null);

    const formData = new FormData();
    formData.append("image", selectedFile);

    try {
      const apiUrl = import.meta.env.VITE_API_URL || "/api";
      const response = await fetch(`${apiUrl}/upload`, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Upload failed");
      }

      const result = await response.json();
      console.log("Upload successful:", result);

      setUploadSuccess(true);
      setSelectedFile(null);
      setPreviewUrl(null);

      const fileInput = document.getElementById(
        "file-input"
      ) as HTMLInputElement;
      if (fileInput) {
        fileInput.value = "";
      }
    } catch (error) {
      console.error("Upload error:", error);
      setUploadError(error instanceof Error ? error.message : "Upload failed");
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-2xl mx-auto px-4">
        <div className="bg-white rounded-lg shadow-md p-8">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-gray-900 mb-2">
              Upload Image
            </h1>
            <p className="text-gray-600">
              Select a JPEG image to upload (max 10MB)
            </p>
          </div>

          <div className="space-y-6">
            {/* file input */}
            <div>
              <label
                htmlFor="file-input"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Choose Image File
              </label>
              <input
                id="file-input"
                type="file"
                accept="image/jpeg,image/jpg"
                onChange={handleFileSelect}
                className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 border border-gray-300 rounded-md"
              />
            </div>

            {/* preview */}
            {previewUrl && (
              <div className="mt-4">
                <p className="text-sm font-medium text-gray-700 mb-2">
                  Preview:
                </p>
                <img
                  src={previewUrl}
                  alt="Preview"
                  className="max-w-full h-auto max-h-64 rounded-md border border-gray-300"
                />
                <p className="text-sm text-gray-500 mt-2">
                  {selectedFile?.name} (
                  {(selectedFile?.size || 0 / (1024 * 1024)).toFixed(2)} MB)
                </p>
              </div>
            )}

            {/* upload */}
            <div>
              <button
                onClick={handleUpload}
                disabled={!selectedFile || uploading}
                className={`w-full py-3 px-4 rounded-md text-white font-semibold ${
                  !selectedFile || uploading
                    ? "bg-gray-400 cursor-not-allowed"
                    : "bg-blue-600 hover:bg-blue-700 active:bg-blue-800"
                } transition-colors`}
              >
                {uploading ? "Uploading..." : "Upload Image"}
              </button>
            </div>

            {/* ok message */}
            {uploadSuccess && (
              <div className="bg-green-50 border border-green-200 rounded-md p-4">
                <div className="flex">
                  <div className="text-green-800">
                    <p className="font-semibold">Upload Successful!</p>
                    <p className="text-sm">
                      Your image has been uploaded successfully.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* error message */}
            {uploadError && (
              <div className="bg-red-50 border border-red-200 rounded-md p-4">
                <div className="flex">
                  <div className="text-red-800">
                    <p className="font-semibold">Upload Failed</p>
                    <p className="text-sm">{uploadError}</p>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* nav */}
          <div className="mt-8 pt-6 border-t border-gray-200">
            <div className="flex justify-between">
              <Link
                to="/"
                className="text-blue-600 hover:text-blue-800 font-medium"
              >
                ← Back to Home
              </Link>
              <Link
                to="/gallery"
                className="text-blue-600 hover:text-blue-800 font-medium"
              >
                View Gallery →
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
