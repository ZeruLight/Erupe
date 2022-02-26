#define max(a, b) ((a) > (b) ? (a) : (b))
max(x, y);
max((x), y);
max(x, (y));
max((x), (y));
max((x, 1), y);
max((x, (1, 3)), ((y, 4), 2));
