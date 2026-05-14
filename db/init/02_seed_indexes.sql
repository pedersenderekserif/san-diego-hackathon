-- Seed the indexes table with the known MRF index IDs from reporting_plans.csv.
-- ingestor_id values correspond to payor_id entries in the index_templates table.

-- Cigna
INSERT INTO indexes (id, ingestor_id) VALUES
    ('41ff13ed-6100-4673-8025-e77c8b292131', '5c7844b1-f8ed-468c-9750-8faa952aa9be');

-- Aetna (multiple subsidiary entities, all share the same Aetna ingestor_id)
INSERT INTO indexes (id, ingestor_id) VALUES
    ('8f60d460-83f1-4b80-b238-482850f7104c', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c'),
    ('a1f88fbe-fce1-4143-b82e-1ee2ad1f693b', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c'),
    ('bec9cd15-2dc2-40f5-b421-9fe8785a905b', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c'),
    ('66c867ee-7be8-4064-a95a-b7925277149c', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c'),
    ('7fce58f6-e776-4285-a078-5e855ad507a3', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c'),
    ('7d803415-3310-4660-80f7-54e5d78a32af', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c'),
    ('d2a2749d-823b-4d6b-836e-4ebf67e57b05', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c'),
    ('bcf12d78-e053-44ab-a8a6-2d52c6e7aa9d', '9b29c0e6-21b4-41b0-bd13-a7d8a4342d4c');
