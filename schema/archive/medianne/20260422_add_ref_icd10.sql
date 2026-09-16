-- ICD-10 reference table seeded from WHO release / Kemenkes-localised version
CREATE TABLE IF NOT EXISTS ref_icd10 (
    code        VARCHAR(10)  PRIMARY KEY,
    display     VARCHAR(500) NOT NULL,
    category    VARCHAR(100),
    created_at  TIMESTAMP    DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ref_icd10_search
    ON ref_icd10 USING gin(to_tsvector('simple', code || ' ' || display));
