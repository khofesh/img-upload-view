import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("upload", "routes/upload.tsx"),
  route("gallery", "routes/gallery.tsx"),
] satisfies RouteConfig;
