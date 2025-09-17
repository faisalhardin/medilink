-- Add short_id column to mdl_mst_journey_point table
-- This migration adds a unique short ID field for journey points

ALTER TABLE mdl_mst_journey_point 
ADD COLUMN short_id VARCHAR(8) UNIQUE;

-- Create index for better performance on short_id lookups
CREATE INDEX idx_mdl_mst_journey_point_short_id ON mdl_mst_journey_point(short_id);

-- Add comment to the column
COMMENT ON COLUMN mdl_mst_journey_point.short_id IS 'Short unique identifier for journey point (Base58 encoded, 8 characters)';

ALTER TABLE mdl_mst_journey_point 
ALTER COLUMN short_id SET NOT NULL;