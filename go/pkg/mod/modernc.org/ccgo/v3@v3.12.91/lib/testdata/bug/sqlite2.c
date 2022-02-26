typedef struct Expr Expr;
typedef struct Table Table;
typedef struct Column Column;

struct Column {
  char *zName;     /* Name of this column, \000, then the type */
  Expr *pDflt;     /* Default value of this column */
  char *zColl;     /* Collating sequence.  If NULL, use the default */
  unsigned char notNull;      /* An OE_ code for handling a NOT NULL constraint */
  char affinity;   /* One of the SQLITE_AFF_... values */
  unsigned char szEst;        /* Estimated size of value in this column. sizeof(INT)==1 */
  unsigned char colFlags;     /* Boolean properties.  See COLFLAG_ defines below */
};

struct Table {
  char *zName;         /* Name of the table or view */
  Column *aCol;        /* Information about each column */
  void /*Index*/ *pIndex;       /* List of SQL indexes on this table. */
  void /*Select*/ *pSelect;     /* NULL for tables.  Points to definition if a view. */
  void /*FKey*/ *pFKey;         /* Linked list of all foreign keys in this table */
  char *zColAff;       /* String defining the affinity of each column */
  void /*ExprList*/ *pCheck;    /* All CHECK constraints */
                       /*   ... also used as column name list in a VIEW */
  int tnum;            /* Root BTree page for this table */
  unsigned nTabRef;         /* Number of pointers to this Table */
  unsigned tabFlags;        /* Mask of TF_* values */
  short iPKey;           /* If not negative, use aCol[iPKey] as the rowid */
  short nCol;            /* Number of columns in this table */
  int /*LogEst*/ nRowLogEst;   /* Estimated rows in table - from sqlite_stat1 table */
  int /*LogEst*/ szTabRow;     /* Estimated size of each table row in bytes */
#ifdef SQLITE_ENABLE_COSTMULT
  int /*LogEst*/ costMult;     /* Cost multiplier for using this table */
#endif
  unsigned char keyConf;          /* What to do in case of uniqueness conflict on iPKey */
#ifndef SQLITE_OMIT_ALTERTABLE
  int addColOffset;    /* Offset in CREATE TABLE stmt to add a new column */
#endif
#ifndef SQLITE_OMIT_VIRTUALTABLE
  int nModuleArg;      /* Number of arguments to the module */
  char **azModuleArg;  /* 0: module 1: schema 2: vtab name 3...: args */
  void /*VTable*/ *pVTable;     /* List of VTable objects. */
#endif
  void /*Trigger*/ *pTrigger;   /* List of triggers stored in pSchema */
  void /*Schema*/ *pSchema;     /* Schema that contains this table */
  Table *pNextZombie;  /* Next on the Parse.pZombieTab list */
};

struct Expr {
  unsigned char op;                 /* Operation performed by this node */
  char affExpr;          /* affinity, or RAISE type */
  unsigned flags;             /* Various flags.  EP_* See below */
  union {
    char *zToken;          /* Token value. Zero terminated and dequoted */
    int iValue;            /* Non-negative integer value if EP_IntValue */
  } u;

  /* If the EP_TokenOnly flag is set in the Expr.flags mask, then no
  ** space is allocated for the fields below this point. An attempt to
  ** access them will result in a segfault or malfunction.
  *********************************************************************/

  Expr *pLeft;           /* Left subnode */
  Expr *pRight;          /* Right subnode */
  union {
    void /*ExprList*/ *pList;     /* op = IN, EXISTS, SELECT, CASE, FUNCTION, BETWEEN */
    void /*Select*/ *pSelect;     /* EP_xIsSelect and op = IN, EXISTS, SELECT */
  } x;

  /* If the EP_Reduced flag is set in the Expr.flags mask, then no
  ** space is allocated for the fields below this point. An attempt to
  ** access them will result in a segfault or malfunction.
  *********************************************************************/

#if SQLITE_MAX_EXPR_DEPTH>0
  int nHeight;           /* Height of the tree headed by this node */
#endif
  int iTable;            /* TK_COLUMN: cursor number of table holding column
                         ** TK_REGISTER: register number
                         ** TK_TRIGGER: 1 -> new, 0 -> old
                         ** EP_Unlikely:  134217728 times likelihood
                         ** TK_IN: ephemerial table holding RHS
                         ** TK_SELECT_COLUMN: Number of columns on the LHS
                         ** TK_SELECT: 1st register of result vector */
  int /*ynVar*/ iColumn;         /* TK_COLUMN: column index.  -1 for rowid.
                         ** TK_VARIABLE: variable number (always >= 1).
                         ** TK_SELECT_COLUMN: column of the result vector */
  short iAgg;              /* Which entry in pAggInfo->aCol[] or ->aFunc[] */
  short iRightJoinTable;   /* If EP_FromJoin, the right table of the join */
  unsigned char op2;                /* TK_REGISTER/TK_TRUTH: original value of Expr.op
                         ** TK_COLUMN: the value of p5 for OP_Column
                         ** TK_AGG_FUNCTION: nesting depth */
  void /*AggInfo*/ *pAggInfo;     /* Used by TK_AGG_COLUMN and TK_AGG_FUNCTION */
  union {
    Table *pTab;           /* TK_COLUMN: Table containing column. Can be NULL
                           ** for a column of an index on an expression */
    void /*Window*/ *pWin;          /* EP_WinFunc: Window/Filter defn for a function */
    struct {               /* TK_IN, TK_SELECT, and TK_EXISTS */
      int iAddr;             /* Subroutine entry address */
      int regReturn;         /* Register used to hold return address */
    } sub;
  } y;
};

void f1(Column c) {
	__builtin_printf("%i: c.zColl %s\n", __LINE__, c.zColl);
}

void f2(Column *cols) {
	__builtin_printf("%i: cols[1].zColl %s\n", __LINE__, cols[1].zColl);
}

Table table;

Table *pTab = &table;

void f3(Table *pTab) {
	__builtin_printf("%i: pTab->aCol[1].zColl %s\n", __LINE__, pTab->aCol[1].zColl);
}

Expr expr;

void f4(Expr *p) {
 	__builtin_printf("%i: p->y.pTab->aCol[1].zColl %s\n", __LINE__, p->y.pTab->aCol[1].zColl);
}

// const char *zColl = p->y.pTab->aCol[j].zColl;

int main() {
	Column c;
	c.zColl = "foo";
	f1(c);
	Column *cols = __builtin_malloc(3*sizeof(Column));
	cols[1] = c;
	f2(cols);
	table.aCol = cols;
	f3(pTab);
	expr.y.pTab = pTab;
	f4(&expr);
}
