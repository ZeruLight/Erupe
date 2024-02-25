BEGIN;

INSERT INTO public.cafebonus (time_req, item_type, item_id, quantity)
VALUES
    (1800, 17, 0, 250),
    (3600, 17, 0, 500),
    (7200, 17, 0, 1000),
    (10800, 17, 0, 1500),
    (18000, 17, 0, 1750),
    (28800, 17, 0, 3000),
    (43200, 17, 0, 4000);

END;