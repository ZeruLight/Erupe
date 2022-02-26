// [0]6.7.7, 7, p.124

typedef void fv(int), (*pfv)(int);

void (*signal(int, void (*)(int)))(int);
fv *signal(int, fv *);
pfv signal(int, pfv);

// Denormalized forms.

void ((*signal(int, void (*)(int)))(int));
void (((*signal(int, void (*)(int)))(int)));

void (*signal(int, void ((*))(int)))(int);
void (*signal(int, void (((*)))(int)))(int);

fv (*signal(int, fv *));
fv ((*signal(int, fv *)));
fv *signal(int, fv (*));
fv *signal(int, fv ((*)));

pfv (signal(int, pfv));
pfv ((signal(int, pfv)));

// ----------------------------------------------------------------------------

typedef int t;

int f29();
t f29();

int *f32();
t *f32();

int (*f35)();
t (*f35)();

int (*f38())();
t (*f38())();

typedef int *t2;

int *f43();
t2 f43();

int *(*f46)();
t2 (*f46)();

int *(*f49());
t2 (*f49());

int **f52();
t2 *f52();

int a55[4];
t a55[4];

int *a58[4];
t *a58[4];

int *a61[4];
t2 a61[4];

int **a64[4];
t2 *a64[4];

int *(*a67[4]);
t2 *(a67[4]);

int (**a67[4]);
t2 (*a67[4]);
