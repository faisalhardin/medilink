-- Add Indonesian localization columns and seed dental ICD-10 codes.
-- display_id: Indonesian display name
-- synonyms_id: Indonesian lay terms for search (e.g., "gigi berlubang" for caries)
--
-- License note: Full ICD-10 redistribution requires a WHO license.
-- Indonesian names (display_id, synonyms_id) should be clinically reviewed.
-- This seeds only a dental subset (K00-K14, S02.5, S03.2, Z01.2, Z46.3, B37.0, B00.2).

-- Add new columns for Indonesian localization
ALTER TABLE ref_icd10 ADD COLUMN IF NOT EXISTS display_id VARCHAR(500);
ALTER TABLE ref_icd10 ADD COLUMN IF NOT EXISTS synonyms_id TEXT[];

-- Drop and recreate index to include new columns for search
DROP INDEX IF EXISTS idx_ref_icd10_search;
CREATE INDEX idx_ref_icd10_search
    ON ref_icd10 USING gin(to_tsvector('simple', code || ' ' || display || ' ' || COALESCE(display_id, '') || ' ' || COALESCE(array_to_string(synonyms_id, ' '), '')));

-- Seed dental ICD-10 codes (WHO ICD-10 2019 titles, Indonesian names for review)
-- Using ON CONFLICT to be idempotent
INSERT INTO ref_icd10 (code, display, display_id, synonyms_id, category) VALUES
-- K00: Disorders of tooth development and eruption
('K00', 'Disorders of tooth development and eruption', 'Gangguan perkembangan dan erupsi gigi', ARRAY['kelainan gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.0', 'Anodontia', 'Anodonsia', ARRAY['tidak tumbuh gigi', 'gigi tidak ada'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.1', 'Supernumerary teeth', 'Gigi supernumerari', ARRAY['gigi lebih', 'gigi berlebih'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.2', 'Abnormalities of size and form of teeth', 'Kelainan ukuran dan bentuk gigi', ARRAY['gigi abnormal'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.3', 'Mottled teeth', 'Gigi berbintik', ARRAY['fluorosis gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.4', 'Disturbances in tooth formation', 'Gangguan pembentukan gigi', ARRAY['hipoplasia email', 'kelainan email'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.5', 'Hereditary disturbances in tooth structure, not elsewhere classified', 'Gangguan herediter struktur gigi', ARRAY['amelogenesis imperfekta', 'dentinogenesis imperfekta'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.6', 'Disturbances in tooth eruption', 'Gangguan erupsi gigi', ARRAY['gigi terlambat tumbuh', 'erupsi terlambat'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.7', 'Teething syndrome', 'Sindrom tumbuh gigi', ARRAY['sakit tumbuh gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K00.8', 'Other disorders of tooth development', 'Gangguan perkembangan gigi lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K00.9', 'Disorder of tooth development, unspecified', 'Gangguan perkembangan gigi tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K01: Embedded and impacted teeth
('K01', 'Embedded and impacted teeth', 'Gigi terpendam dan impaksi', ARRAY['gigi impaksi', 'gigi bungsu'], 'Diseases of oral cavity, salivary glands and jaws'),
('K01.0', 'Embedded teeth', 'Gigi terpendam', ARRAY['gigi tertanam'], 'Diseases of oral cavity, salivary glands and jaws'),
('K01.1', 'Impacted teeth', 'Gigi impaksi', ARRAY['gigi bungsu impaksi', 'gigi miring', 'impaksi molar tiga'], 'Diseases of oral cavity, salivary glands and jaws'),

-- K02: Dental caries
('K02', 'Dental caries', 'Karies gigi', ARRAY['gigi berlubang', 'karies', 'lubang gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K02.0', 'Caries limited to enamel', 'Karies terbatas pada email', ARRAY['karies email', 'lubang email'], 'Diseases of oral cavity, salivary glands and jaws'),
('K02.1', 'Caries of dentine', 'Karies dentin', ARRAY['karies dentin', 'lubang dalam', 'gigi berlubang dalam'], 'Diseases of oral cavity, salivary glands and jaws'),
('K02.2', 'Caries of cementum', 'Karies sementum', ARRAY['karies akar'], 'Diseases of oral cavity, salivary glands and jaws'),
('K02.3', 'Arrested dental caries', 'Karies gigi terhenti', ARRAY['karies terhenti'], 'Diseases of oral cavity, salivary glands and jaws'),
('K02.4', 'Odontoclasia', 'Odontoklasia', ARRAY['resorpsi gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K02.5', 'Caries with exposure of pulp', 'Karies dengan terbukanya pulpa', ARRAY['karies pulpa', 'gigi bolong sampai saraf'], 'Diseases of oral cavity, salivary glands and jaws'),
('K02.8', 'Other dental caries', 'Karies gigi lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K02.9', 'Dental caries, unspecified', 'Karies gigi tidak spesifik', ARRAY['gigi berlubang'], 'Diseases of oral cavity, salivary glands and jaws'),

-- K03: Other diseases of hard tissues of teeth
('K03', 'Other diseases of hard tissues of teeth', 'Penyakit jaringan keras gigi lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K03.0', 'Excessive attrition of teeth', 'Atrisi gigi berlebihan', ARRAY['gigi aus'], 'Diseases of oral cavity, salivary glands and jaws'),
('K03.1', 'Abrasion of teeth', 'Abrasi gigi', ARRAY['gigi terkikis'], 'Diseases of oral cavity, salivary glands and jaws'),
('K03.2', 'Erosion of teeth', 'Erosi gigi', ARRAY['pengikisan gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K03.3', 'Pathological resorption of teeth', 'Resorpsi patologis gigi', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K03.4', 'Hypercementosis', 'Hipersementosis', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K03.5', 'Ankylosis of teeth', 'Ankilosis gigi', ARRAY['gigi menyatu tulang'], 'Diseases of oral cavity, salivary glands and jaws'),
('K03.6', 'Deposits [accretions] on teeth', 'Deposit pada gigi', ARRAY['karang gigi', 'kalkulus', 'plak gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K03.7', 'Posteruptive colour changes of dental hard tissues', 'Perubahan warna gigi setelah erupsi', ARRAY['gigi kuning', 'gigi berubah warna', 'diskolorasi gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K03.8', 'Other specified diseases of hard tissues of teeth', 'Penyakit jaringan keras gigi spesifik lainnya', ARRAY['gigi sensitif', 'gigi ngilu'], 'Diseases of oral cavity, salivary glands and jaws'),
('K03.9', 'Disease of hard tissues of teeth, unspecified', 'Penyakit jaringan keras gigi tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K04: Diseases of pulp and periapical tissues
('K04', 'Diseases of pulp and periapical tissues', 'Penyakit pulpa dan jaringan periapikal', ARRAY['sakit saraf gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.0', 'Pulpitis', 'Pulpitis', ARRAY['radang pulpa', 'saraf gigi meradang', 'sakit gigi berdenyut'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.1', 'Necrosis of pulp', 'Nekrosis pulpa', ARRAY['pulpa mati', 'saraf gigi mati', 'gigi mati'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.2', 'Pulp degeneration', 'Degenerasi pulpa', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K04.3', 'Abnormal hard tissue formation in pulp', 'Pembentukan jaringan keras abnormal dalam pulpa', ARRAY['kalsifikasi pulpa'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.4', 'Acute apical periodontitis of pulpal origin', 'Periodontitis apikal akut asal pulpa', ARRAY['abses akut', 'infeksi akar gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.5', 'Chronic apical periodontitis', 'Periodontitis apikal kronis', ARRAY['granuloma', 'infeksi akar kronis'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.6', 'Periapical abscess with sinus', 'Abses periapikal dengan sinus', ARRAY['abses gigi dengan fistula', 'bisul gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.7', 'Periapical abscess without sinus', 'Abses periapikal tanpa sinus', ARRAY['abses gigi', 'bengkak gigi', 'nanah gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.8', 'Radicular cyst', 'Kista radikuler', ARRAY['kista akar gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K04.9', 'Other and unspecified diseases of pulp and periapical tissues', 'Penyakit pulpa dan periapikal lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K05: Gingivitis and periodontal diseases
('K05', 'Gingivitis and periodontal diseases', 'Gingivitis dan penyakit periodontal', ARRAY['radang gusi', 'penyakit gusi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K05.0', 'Acute gingivitis', 'Gingivitis akut', ARRAY['radang gusi akut', 'gusi merah bengkak'], 'Diseases of oral cavity, salivary glands and jaws'),
('K05.1', 'Chronic gingivitis', 'Gingivitis kronis', ARRAY['radang gusi kronis', 'gusi berdarah', 'gingivitis'], 'Diseases of oral cavity, salivary glands and jaws'),
('K05.2', 'Acute periodontitis', 'Periodontitis akut', ARRAY['radang jaringan penyangga gigi akut'], 'Diseases of oral cavity, salivary glands and jaws'),
('K05.3', 'Chronic periodontitis', 'Periodontitis kronis', ARRAY['gigi goyang', 'pyorrhea', 'periodontitis'], 'Diseases of oral cavity, salivary glands and jaws'),
('K05.4', 'Periodontosis', 'Periodontosis', ARRAY['kerusakan tulang gigi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K05.5', 'Other periodontal diseases', 'Penyakit periodontal lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K05.6', 'Periodontal disease, unspecified', 'Penyakit periodontal tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K06: Other disorders of gingiva and edentulous alveolar ridge
('K06', 'Other disorders of gingiva and edentulous alveolar ridge', 'Gangguan gusi dan ridge alveolar edentulus lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K06.0', 'Gingival recession', 'Resesi gingiva', ARRAY['gusi turun', 'gusi menyusut', 'akar gigi terlihat'], 'Diseases of oral cavity, salivary glands and jaws'),
('K06.1', 'Gingival enlargement', 'Pembesaran gingiva', ARRAY['gusi bengkak', 'hiperplasia gusi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K06.2', 'Gingival and edentulous alveolar ridge lesions associated with trauma', 'Lesi gusi dan ridge alveolar edentulus terkait trauma', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K06.8', 'Other specified disorders of gingiva and edentulous alveolar ridge', 'Gangguan gusi spesifik lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K06.9', 'Disorder of gingiva and edentulous alveolar ridge, unspecified', 'Gangguan gusi tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K07: Dentofacial anomalies [including malocclusion]
('K07', 'Dentofacial anomalies [including malocclusion]', 'Anomali dentofasial [termasuk maloklusi]', ARRAY['gigi tidak rata', 'maloklusi'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.0', 'Major anomalies of jaw size', 'Anomali besar ukuran rahang', ARRAY['rahang besar', 'rahang kecil'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.1', 'Anomalies of jaw-cranial base relationship', 'Anomali hubungan rahang-basis kranial', ARRAY['prognati', 'retrognati'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.2', 'Anomalies of dental arch relationship', 'Anomali hubungan lengkung gigi', ARRAY['crossbite', 'overbite', 'underbite'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.3', 'Anomalies of tooth position', 'Anomali posisi gigi', ARRAY['gigi berjejal', 'gigi miring', 'gigi tonggos'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.4', 'Malocclusion, unspecified', 'Maloklusi tidak spesifik', ARRAY['gigitan tidak normal'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.5', 'Dentofacial functional abnormalities', 'Abnormalitas fungsional dentofasial', ARRAY['rahang bunyi', 'TMJ'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.6', 'Temporomandibular joint disorders', 'Gangguan sendi temporomandibula', ARRAY['TMJ', 'rahang sakit', 'rahang bunyi klik', 'susah buka mulut'], 'Diseases of oral cavity, salivary glands and jaws'),
('K07.8', 'Other dentofacial anomalies', 'Anomali dentofasial lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K07.9', 'Dentofacial anomaly, unspecified', 'Anomali dentofasial tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K08: Other disorders of teeth and supporting structures
('K08', 'Other disorders of teeth and supporting structures', 'Gangguan gigi dan struktur penyangga lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K08.0', 'Exfoliation of teeth due to systemic causes', 'Eksfoliasi gigi karena penyebab sistemik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K08.1', 'Loss of teeth due to accident, extraction or local periodontal disease', 'Kehilangan gigi karena kecelakaan, pencabutan atau penyakit periodontal lokal', ARRAY['gigi copot', 'gigi tanggal', 'ompong'], 'Diseases of oral cavity, salivary glands and jaws'),
('K08.2', 'Atrophy of edentulous alveolar ridge', 'Atrofi ridge alveolar edentulus', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K08.3', 'Retained dental root', 'Sisa akar gigi', ARRAY['akar gigi tertinggal', 'sisa akar'], 'Diseases of oral cavity, salivary glands and jaws'),
('K08.8', 'Other specified disorders of teeth and supporting structures', 'Gangguan gigi dan struktur penyangga spesifik lainnya', ARRAY['gigi retak', 'gigi patah'], 'Diseases of oral cavity, salivary glands and jaws'),
('K08.9', 'Disorder of teeth and supporting structures, unspecified', 'Gangguan gigi dan struktur penyangga tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K09: Cysts of oral region, not elsewhere classified
('K09', 'Cysts of oral region, not elsewhere classified', 'Kista rongga mulut', ARRAY['kista mulut'], 'Diseases of oral cavity, salivary glands and jaws'),
('K09.0', 'Developmental odontogenic cysts', 'Kista odontogenik developmental', ARRAY['kista dentigerous', 'kista folikuler'], 'Diseases of oral cavity, salivary glands and jaws'),
('K09.1', 'Developmental (nonodontogenic) cysts of oral region', 'Kista developmental non-odontogenik rongga mulut', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K09.2', 'Other cysts of jaw', 'Kista rahang lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K09.8', 'Other cysts of oral region, not elsewhere classified', 'Kista rongga mulut lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K09.9', 'Cyst of oral region, unspecified', 'Kista rongga mulut tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K10: Other diseases of jaws
('K10', 'Other diseases of jaws', 'Penyakit rahang lainnya', ARRAY['sakit rahang'], 'Diseases of oral cavity, salivary glands and jaws'),
('K10.0', 'Developmental disorders of jaws', 'Gangguan perkembangan rahang', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K10.1', 'Giant cell granuloma, central', 'Granuloma sel raksasa, sentral', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K10.2', 'Inflammatory conditions of jaws', 'Kondisi inflamasi rahang', ARRAY['osteomielitis rahang', 'infeksi tulang rahang'], 'Diseases of oral cavity, salivary glands and jaws'),
('K10.3', 'Alveolitis of jaws', 'Alveolitis rahang', ARRAY['dry socket', 'alveolitis'], 'Diseases of oral cavity, salivary glands and jaws'),
('K10.8', 'Other specified diseases of jaws', 'Penyakit rahang spesifik lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K10.9', 'Disease of jaws, unspecified', 'Penyakit rahang tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K11: Diseases of salivary glands
('K11', 'Diseases of salivary glands', 'Penyakit kelenjar ludah', ARRAY['kelenjar ludah'], 'Diseases of oral cavity, salivary glands and jaws'),
('K11.0', 'Atrophy of salivary gland', 'Atrofi kelenjar ludah', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K11.1', 'Hypertrophy of salivary gland', 'Hipertrofi kelenjar ludah', ARRAY['kelenjar ludah bengkak'], 'Diseases of oral cavity, salivary glands and jaws'),
('K11.2', 'Sialadenitis', 'Sialadenitis', ARRAY['radang kelenjar ludah', 'infeksi kelenjar ludah'], 'Diseases of oral cavity, salivary glands and jaws'),
('K11.3', 'Abscess of salivary gland', 'Abses kelenjar ludah', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K11.4', 'Fistula of salivary gland', 'Fistula kelenjar ludah', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K11.5', 'Sialolithiasis', 'Sialolitiasis', ARRAY['batu kelenjar ludah', 'batu ludah'], 'Diseases of oral cavity, salivary glands and jaws'),
('K11.6', 'Mucocele of salivary gland', 'Mukokel kelenjar ludah', ARRAY['mukokel', 'kista ludah'], 'Diseases of oral cavity, salivary glands and jaws'),
('K11.7', 'Disturbances of salivary secretion', 'Gangguan sekresi ludah', ARRAY['mulut kering', 'xerostomia', 'air liur berlebih'], 'Diseases of oral cavity, salivary glands and jaws'),
('K11.8', 'Other diseases of salivary glands', 'Penyakit kelenjar ludah lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K11.9', 'Disease of salivary gland, unspecified', 'Penyakit kelenjar ludah tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K12: Stomatitis and related lesions
('K12', 'Stomatitis and related lesions', 'Stomatitis dan lesi terkait', ARRAY['radang mulut', 'sariawan'], 'Diseases of oral cavity, salivary glands and jaws'),
('K12.0', 'Recurrent oral aphthae', 'Aftosa oral rekuren', ARRAY['sariawan', 'stomatitis aftosa', 'luka mulut'], 'Diseases of oral cavity, salivary glands and jaws'),
('K12.1', 'Other forms of stomatitis', 'Bentuk stomatitis lainnya', ARRAY['radang mulut'], 'Diseases of oral cavity, salivary glands and jaws'),
('K12.2', 'Cellulitis and abscess of mouth', 'Selulitis dan abses mulut', ARRAY['infeksi mulut', 'abses mulut'], 'Diseases of oral cavity, salivary glands and jaws'),
('K12.3', 'Oral mucositis (ulcerative)', 'Mukositis oral (ulseratif)', ARRAY['luka mulut kemoterapi'], 'Diseases of oral cavity, salivary glands and jaws'),

-- K13: Other diseases of lip and oral mucosa
('K13', 'Other diseases of lip and oral mucosa', 'Penyakit bibir dan mukosa mulut lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K13.0', 'Diseases of lips', 'Penyakit bibir', ARRAY['bibir pecah', 'cheilitis'], 'Diseases of oral cavity, salivary glands and jaws'),
('K13.1', 'Cheek and lip biting', 'Gigitan pipi dan bibir', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K13.2', 'Leukoplakia and other disturbances of oral epithelium, including tongue', 'Leukoplakia dan gangguan epitel mulut lainnya', ARRAY['bercak putih mulut', 'leukoplakia'], 'Diseases of oral cavity, salivary glands and jaws'),
('K13.3', 'Hairy leukoplakia', 'Leukoplakia berbulu', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K13.4', 'Granuloma and granuloma-like lesions of oral mucosa', 'Granuloma dan lesi mirip granuloma mukosa mulut', ARRAY['pyogenic granuloma'], 'Diseases of oral cavity, salivary glands and jaws'),
('K13.5', 'Oral submucous fibrosis', 'Fibrosis submukosa oral', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K13.6', 'Irritative hyperplasia of oral mucosa', 'Hiperplasia iritatif mukosa mulut', ARRAY['epulis'], 'Diseases of oral cavity, salivary glands and jaws'),
('K13.7', 'Other and unspecified lesions of oral mucosa', 'Lesi mukosa mulut lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- K14: Diseases of tongue
('K14', 'Diseases of tongue', 'Penyakit lidah', ARRAY['sakit lidah'], 'Diseases of oral cavity, salivary glands and jaws'),
('K14.0', 'Glossitis', 'Glositis', ARRAY['radang lidah', 'lidah meradang'], 'Diseases of oral cavity, salivary glands and jaws'),
('K14.1', 'Geographic tongue', 'Lidah geografik', ARRAY['lidah peta'], 'Diseases of oral cavity, salivary glands and jaws'),
('K14.2', 'Median rhomboid glossitis', 'Glositis romboid median', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K14.3', 'Hypertrophy of tongue papillae', 'Hipertrofi papila lidah', ARRAY['lidah berbulu', 'lidah hitam'], 'Diseases of oral cavity, salivary glands and jaws'),
('K14.4', 'Atrophy of tongue papillae', 'Atrofi papila lidah', ARRAY['lidah licin'], 'Diseases of oral cavity, salivary glands and jaws'),
('K14.5', 'Plicated tongue', 'Lidah berkerut', ARRAY['fissured tongue', 'lidah retak'], 'Diseases of oral cavity, salivary glands and jaws'),
('K14.6', 'Glossodynia', 'Glosodinnia', ARRAY['lidah perih', 'burning tongue', 'lidah panas'], 'Diseases of oral cavity, salivary glands and jaws'),
('K14.8', 'Other diseases of tongue', 'Penyakit lidah lainnya', NULL, 'Diseases of oral cavity, salivary glands and jaws'),
('K14.9', 'Disease of tongue, unspecified', 'Penyakit lidah tidak spesifik', NULL, 'Diseases of oral cavity, salivary glands and jaws'),

-- Additional requested codes
('S02.5', 'Fracture of tooth', 'Fraktur gigi', ARRAY['gigi patah', 'gigi retak', 'gigi pecah'], 'Injuries to the head'),
('S03.2', 'Dislocation of tooth', 'Dislokasi gigi', ARRAY['gigi lepas', 'gigi copot trauma', 'avulsi gigi', 'luksasi gigi'], 'Injuries to the head'),
('Z01.2', 'Dental examination', 'Pemeriksaan gigi', ARRAY['kontrol gigi', 'check up gigi', 'periksa gigi'], 'Factors influencing health status and contact with health services'),
('Z46.3', 'Fitting and adjustment of dental prosthetic device', 'Pemasangan dan penyesuaian alat prostetik gigi', ARRAY['pasang gigi palsu', 'gigi tiruan'], 'Factors influencing health status and contact with health services'),
('B37.0', 'Candidal stomatitis', 'Stomatitis kandida', ARRAY['oral thrush', 'jamur mulut', 'kandidiasis oral'], 'Mycoses'),
('B00.2', 'Herpesviral gingivostomatitis and pharyngotonsillitis', 'Gingivostomatitis dan faringotonsillitis herpesviral', ARRAY['herpes mulut', 'herpes gusi'], 'Viral infections characterized by skin and mucous membrane lesions')

ON CONFLICT (code) DO UPDATE SET
    display = EXCLUDED.display,
    display_id = EXCLUDED.display_id,
    synonyms_id = EXCLUDED.synonyms_id,
    category = EXCLUDED.category;
