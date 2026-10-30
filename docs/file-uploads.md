# Chat file attachments

The authenticated `POST /api/files/upload` endpoint accepts a multipart form with one `file`, a chat `type` (`dm` or `channel`), the matching `receiverId` or `channelId`, and an optional `caption`. Files may be up to 50 MiB. The response contains the created chat message; it is also delivered to connected chat members over the existing WebSocket channel.

Set `MINIO_ENDPOINT`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, and `MINIO_BUCKET` in the backend environment. Set `MINIO_SECURE=true` when the configured endpoint uses HTTPS (it defaults to false for local MinIO). The backend creates the bucket on startup when it does not exist. Keep the bucket private. The configured endpoint used for signed downloads must be reachable by the browser; for local development use the host-accessible MinIO address.

Attachment object keys and metadata are persisted on the existing `messages` table. Run the normal backend migration command (`go run . migrate`) after deploying the model update. `GET /api/files/{messageID}/download` requires authentication and chat membership, then redirects to a five-minute signed MinIO URL with download disposition. Image attachments also include a separate authorized `/preview` route that serves supported raster image formats inline; SVG is kept as a downloadable file.
