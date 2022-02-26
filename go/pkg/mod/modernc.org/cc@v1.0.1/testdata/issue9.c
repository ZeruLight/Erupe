typedef short JBLOCK[64];
typedef JBLOCK *JBLOCKROW;
typedef JBLOCKROW *JBLOCKARRAY;
typedef JBLOCKARRAY *JBLOCKIMAGE;

short v6[64];
short (v6)[64];
JBLOCK v6;
JBLOCK (v6);

short (*v11)[64];
short ((*v11))[64];
short ((*v11)[64]);
JBLOCKROW v11;
JBLOCKROW (v11);

short (**v17)[64];
short ((**v17))[64];
short ((**v17)[64]);
JBLOCKARRAY v17;
JBLOCKARRAY (v17);

short (***v23)[64];
short ((***v23))[64];
short ((***v23)[64]);
JBLOCKIMAGE v23;
JBLOCKIMAGE (v23);
