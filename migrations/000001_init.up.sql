CREATE TABLE "spotify_connect" (
    "owner_id"      uuid                     PRIMARY KEY,
    "refresh_token" text,
    "device_id"     text,
    "gmail"         text,
    "created_at"    timestamp with time zone
);
