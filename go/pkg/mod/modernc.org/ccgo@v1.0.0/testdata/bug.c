struct sqlite3 {
  int mutex;
};
typedef struct sqlite3 sqlite3;

struct Mem {
  sqlite3 *db;        /* The associated database connection */
};

typedef struct Mem sqlite3_value;

sqlite3_value Val;
sqlite3 db;



int main() {
	db.mutex = 42;
	Val.db = &db;
	sqlite3_value* pVal = &Val;
	if (pVal->db->mutex != 42)
		abort();
}
