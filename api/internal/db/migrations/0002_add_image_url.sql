-- image_url: populated from a feed's item.Image.URL when present (see
-- feeds.ExtractItems). Empty/NULL when a feed provides no image - the API
-- layer substitutes a placeholder for these rather than leaving it blank,
-- so every client gets consistent behavior for free.
ALTER TABLE posts ADD COLUMN image_url TEXT;
