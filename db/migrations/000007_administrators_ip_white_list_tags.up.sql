-- administrators new fields

ALTER TABLE administrators ADD COLUMN ip_white_list TEXT;
ALTER TABLE administrators ADD COLUMN allowed_tags TEXT[];
