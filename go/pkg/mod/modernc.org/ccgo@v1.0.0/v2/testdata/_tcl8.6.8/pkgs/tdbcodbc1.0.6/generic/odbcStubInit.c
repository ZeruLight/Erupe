/*
 * odbcStubInit.c --
 *
 *	Stubs tables for the foreign ODBC libraries so that
 *	Tcl extensions can use them without the linker's knowing about them.
 *
 * @CREATED@ 2017-06-05 16:16:37Z by genExtStubs.tcl from ../generic/odbcStubDefs.txt
 *
 * Copyright (c) 2010 by Kevin B. Kenny.
 *
 * Please refer to the file, 'license.terms' for the conditions on
 * redistribution of this file and for a DISCLAIMER OF ALL WARRANTIES.
 *
 *-----------------------------------------------------------------------------
 */

#include <tcl.h>
#ifdef _WIN32
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#endif
#include "fakesql.h"

/*
 * Static data used in this file
 */

/*
 * Names of the libraries that might contain the ODBC API
 */


/* Uncomment or -DTDBC_NEW_LOADER=1 to use the new loader */
/*#define TDBC_NEW_LOADER 1*/


#ifdef TDBC_NEW_LOADER

#include <stdlib.h>

/* Sorted by name asc. */
static const char *const odbcStubLibNames[] = {
    "odbc", "odbc32", NULL
};
/* Sorted by num desc. No leading dots. Empty first. */
static const char *const odbcStubLibNumbers[] = {
	"", "1.2", "0.1", NULL
};
/* Sorted by name asc. */
static const char *const odbcOptLibNames[] = {
    "odbccp", "odbccp32", "odbcinst", NULL
};
/* Sorted by num desc. No leading dots. Empty first. */
static const char *const odbcOptLibNumbers[] = {
	"", "2.6", "0.0", NULL
};

#else

static const char *const odbcStubLibNames[] = {
    /* @LIBNAMES@: DO NOT EDIT THESE NAMES */
    "odbc32", "odbc", "libodbc32", "libodbc", NULL
    /* @END@ */
};
static const char *const odbcOptLibNames[] = {
    "odbccp", "odbccp32", "odbcinst",
    "libodbccp", "libodbccp32", "libodbcinst", NULL
};

#endif


/*
 * Names of the functions that we need from ODBC
 */

static const char *const odbcSymbolNames[] = {
    /* @SYMNAMES@: DO NOT EDIT THESE NAMES */
    "SQLAllocHandle",
    "SQLBindParameter",
    "SQLCloseCursor",
    "SQLColumnsW",
    "SQLDataSourcesW",
    "SQLDescribeColW",
    "SQLDescribeParam",
    "SQLDisconnect",
    "SQLDriverConnectW",
    "SQLDriversW",
    "SQLEndTran",
    "SQLExecute",
    "SQLFetch",
    "SQLForeignKeysW",
    "SQLFreeHandle",
    "SQLGetConnectAttr",
    "SQLGetData",
    "SQLGetDiagFieldA",
    "SQLGetDiagRecW",
    "SQLGetTypeInfo",
    "SQLMoreResults",
    "SQLNumParams",
    "SQLNumResultCols",
    "SQLPrepareW",
    "SQLPrimaryKeysW",
    "SQLRowCount",
    "SQLSetConnectAttr",
    "SQLSetConnectOption",
    "SQLSetEnvAttr",
    "SQLTablesW",
    NULL
    /* @END@ */
};

/*
 * Table containing pointers to the functions named above.
 */

static odbcStubDefs odbcStubsTable;
const odbcStubDefs* odbcStubs = &odbcStubsTable;

/*
 * Pointers to optional functions in ODBCINST
 */

BOOL (INSTAPI* SQLConfigDataSourceW)(HWND, WORD, LPCWSTR, LPCWSTR)
= NULL;
BOOL (INSTAPI* SQLConfigDataSource)(HWND, WORD, LPCSTR, LPCSTR)
= NULL;
BOOL (INSTAPI* SQLInstallerError)(WORD, DWORD*, LPSTR, WORD, WORD*)
= NULL;


#ifdef TDBC_NEW_LOADER

#ifndef TCL_SHLIB_EXT
#  define TCL_SHLIB_EXT ".so"
#endif

#ifndef LIBPREFIX
#  ifdef __CYGWIN__
#    define LIBPREFIX "cyg"
#  else
#    define LIBPREFIX "lib"
#  endif
#endif

#ifdef __CYGWIN__
#  define TDBC_SHLIB_SEP "-"
#else
#  define TDBC_SHLIB_SEP "."
#endif

const char *const tdbcLibFormats[] = {
	LIBPREFIX "%s" TCL_SHLIB_EXT "%s" "%s",
	"%s" TCL_SHLIB_EXT "%s" "%s",
	NULL
};

/*
 *-----------------------------------------------------------------------------
 *
 * tdbcLoadLib --
 *
 *	Tries to load a shared library given a list of lib names and/or
 *	all combinations of LIBPREFIX, no LIBPREFIX, lib names and lib numbers.
 *	Takes CYGWIN into account.
 *
 * Results:
 *	Returns the handle to the loaded ODBC client library and leaves the
 *	name of the the loaded ODBC client library in the interpreter, or NULL
 *	if the load is unsuccessful and leaves a list of error message(s) in the
 *	interpreter.
 *
 *-----------------------------------------------------------------------------
 */

static Tcl_LoadHandle
tdbcLoadLib (
    Tcl_Interp *interp,			/* Receives errors or lib name if successful */
    const char *const soNames[],	/* Lib names. Kwazy lookup not done if NULL */
    const char *const soNumbers[],	/* Lib numbers for kwazy lookup */
    const char *const soSymbolNames[],	/* Passed to Tcl_LoadFile */
    const void *soStubDefs,		/* Passed to Tcl_LoadFile */
    const char *soList,			/* Lib name list to try first if not NULL */
    const char *const soFormats[]	/* Lib name printf formats for kwazy lookup.
					   Default used if NULL.
					   Maybe not useful as a parameter. */
) {
    const char *const *nam;		/* Name */
    const char *const *num;		/* Number */
    const char *const *fmt;		/* Format */
    Tcl_Obj *lib;			/* Holds lib name during kwazy lookup */
    Tcl_Obj *result;			/* List of errors or lib name if successful */
    Tcl_LoadHandle handle;		/* NULL or handle to loaded lib if successful */

    result = NULL; /* Important! This is eventually returned. */
    handle = NULL; /* Important! This is eventually returned. */

    /*
     * Try to load a lib from a string claiming to be a list of libs?
     */
    if (soList != NULL) {	/* Yes! */
	Tcl_Obj *l;
	Tcl_Obj **els;
	int nels;
	int i;

	/*
         * Make list from string.
	 * Caller is responsible for listability and utf8ness.
         */
	l = Tcl_NewStringObj(soList, -1);
	Tcl_IncrRefCount(l);

	if (Tcl_ListObjGetElements(interp, l, &nels, &els) != TCL_OK) {
	    Tcl_DecrRefCount(l);
	    return NULL;
	}

	result = Tcl_NewListObj(0, NULL);
	Tcl_IncrRefCount(result);

	/*
	 * Left-to-right, trying to load a lib at each iteration.
         */
	for (i = 0; i < nels; i++) {
	    if (Tcl_LoadFile(interp, els[i], soSymbolNames, 0, (void *) soStubDefs, &handle) == TCL_OK) {
		/* Lib found and loaded. Cleanup and setup result. */
		Tcl_DecrRefCount(result); /* Throw away any errors collected. */
		result = Tcl_DuplicateObj(els[i]);
		Tcl_IncrRefCount(result);
		break;
	    }
	    Tcl_ListObjAppendElement(NULL, result, Tcl_GetObjResult(interp)); /* Collect error. */
	    handle = NULL; /* Important! This is eventually returned. */
	}
	Tcl_DecrRefCount(l);
	if (handle != NULL) {
	    goto loadDone;
	}
    }

    /*
     * At this point no lib list was provided or no lib was found in the list.
     */

    /*
     * Done if names not provided.
     */
    if (soNames == NULL) {
	goto loadDone;
    }

    /*
     * Use default format(s) if not supplied.
     */
    if (soFormats == NULL) {
	soFormats = tdbcLibFormats;
    }

    if (result == NULL) {
	result = Tcl_NewListObj(0, NULL);
	Tcl_IncrRefCount(result);
    }

    /*
     * Try every possible combination (aka Kwazy Lookup).
     */
    for (nam =   &soNames[0]; *nam != NULL; nam++) {
    for (num = &soNumbers[0]; *num != NULL; num++) {
    for (fmt = &soFormats[0]; *fmt != NULL; fmt++) {
	lib = Tcl_ObjPrintf(*fmt, *nam, (*num[0] == '\0' ? "" : TDBC_SHLIB_SEP), *num);
	Tcl_IncrRefCount(lib);
	if (Tcl_LoadFile(interp, lib, soSymbolNames, 0, (void *) soStubDefs, &handle) == TCL_OK) {
	    /* Lib found and loaded. Cleanup and setup result. */
	    Tcl_DecrRefCount(result); /* Throw away any errors collected. */
	    result = lib;
	    goto loadDone;
	}
	Tcl_ListObjAppendElement(NULL, result, Tcl_GetObjResult(interp)); /* Collect error. */
	Tcl_DecrRefCount(lib);
	handle = NULL; /* Important! This is eventually returned. */
    }}}

loadDone:

    if (result != NULL) {
	Tcl_SetObjResult(interp, result);
	Tcl_DecrRefCount(result);
    }

    return handle; /* Like I said... */
}

/*
 *-----------------------------------------------------------------------------
 *
 * OdbcInitStubs --
 *
 *	Initialize the Stubs table for the ODBC API
 *
 * Results:
 *	Returns the handle to the loaded ODBC client library, or NULL
 *	if the load is unsuccessful. Leaves an error message in the
 *	interpreter.
 *
 *-----------------------------------------------------------------------------
 */

MODULE_SCOPE Tcl_LoadHandle
OdbcInitStubs(Tcl_Interp* interp,
				/* Tcl interpreter */
	      Tcl_LoadHandle* handle2Ptr)
				/* Pointer to a second load handle
				 * that represents the ODBCINST library */
{
    Tcl_LoadHandle handle;	/* Handle to a load module */
    int status;			/* Status of Tcl library calls for ODBC lib */
    int status2;		/* Status of Tcl library calls for ODBCINST lib */

    SQLConfigDataSourceW = NULL;/* Symbols maybe in ODBCINST lib */
    SQLConfigDataSource = NULL;
    SQLInstallerError = NULL;

    /*
     * Try to load a client library and resolve the ODBC API within it.
     */

    handle = tdbcLoadLib(interp,
	odbcStubLibNames, odbcStubLibNumbers,
	odbcSymbolNames, odbcStubs,
	getenv("TDBC_ODBC_ODBCLIBS"), NULL
    );
    status = (handle == NULL ? TCL_ERROR : TCL_OK);

    /*
     * We've run out of library names (in which case status==TCL_ERROR
     * and the error message reflects the last unsuccessful load attempt).
     */

    if (status != TCL_OK) {
	return NULL;
    }

    /*
     * If a client library is found, then try to load ODBCINST as well.
     */

    *handle2Ptr = tdbcLoadLib(interp,
	odbcOptLibNames, odbcOptLibNumbers,
	NULL, NULL,
	getenv("TDBC_ODBC_ODBCINSTLIBS"), NULL
    );
    status2 = (*handle2Ptr == NULL ? TCL_ERROR : TCL_OK);

    if (status2 == TCL_OK) {
	SQLConfigDataSourceW =
	    (BOOL (INSTAPI*)(HWND, WORD, LPCWSTR, LPCWSTR))
	    Tcl_FindSymbol(NULL, *handle2Ptr, "SQLConfigDataSourceW");
	if (SQLConfigDataSourceW == NULL) {
	    SQLConfigDataSource =
		(BOOL (INSTAPI*)(HWND, WORD, LPCSTR, LPCSTR))
		Tcl_FindSymbol(NULL, *handle2Ptr,
			       "SQLConfigDataSource");
	}
	SQLInstallerError =
	    (BOOL (INSTAPI*)(WORD, DWORD*, LPSTR, WORD, WORD*))
	    Tcl_FindSymbol(NULL, *handle2Ptr, "SQLInstallerError");
    } else {
	Tcl_ResetResult(interp);
    }

    /*
     * We've successfully loaded a library.
     */

    return handle;
}

#else

/*
 *-----------------------------------------------------------------------------
 *
 * OdbcInitStubs --
 *
 *	Initialize the Stubs table for the ODBC API
 *
 * Results:
 *	Returns the handle to the loaded ODBC client library, or NULL
 *	if the load is unsuccessful. Leaves an error message in the
 *	interpreter.
 *
 *-----------------------------------------------------------------------------
 */

MODULE_SCOPE Tcl_LoadHandle
OdbcInitStubs(Tcl_Interp* interp,
				/* Tcl interpreter */
	      Tcl_LoadHandle* handle2Ptr)
				/* Pointer to a second load handle
				 * that represents the ODBCINST library */
{
    int i;
    int status;			/* Status of Tcl library calls */
    Tcl_Obj* path;		/* Path name of a module to be loaded */
    Tcl_Obj* shlibext;		/* Extension to use for load modules */
    Tcl_LoadHandle handle = NULL;
				/* Handle to a load module */

    SQLConfigDataSourceW = NULL;
    SQLConfigDataSource = NULL;
    SQLInstallerError = NULL;

    /*
     * Determine the shared library extension
     */
    status = Tcl_EvalEx(interp, "::info sharedlibextension", -1,
			TCL_EVAL_GLOBAL);
    if (status != TCL_OK) return NULL;
    shlibext = Tcl_GetObjResult(interp);
    Tcl_IncrRefCount(shlibext);

    /*
     * Walk the list of possible library names to find an ODBC client
     */
    status = TCL_ERROR;
    for (i = 0; status == TCL_ERROR && odbcStubLibNames[i] != NULL; ++i) {
	path = Tcl_NewStringObj(odbcStubLibNames[i], -1);
	Tcl_AppendObjToObj(path, shlibext);
	Tcl_IncrRefCount(path);
	Tcl_ResetResult(interp);

	/*
	 * Try to load a client library and resolve the ODBC API within it.
	 */
	status = Tcl_LoadFile(interp, path, odbcSymbolNames, 0,
			      (void*)odbcStubs, &handle);
	Tcl_DecrRefCount(path);
    }

    /*
     * If a client library is found, then try to load ODBCINST as well.
     */
    if (status == TCL_OK) {
	int status2 = TCL_ERROR;
	for (i = 0; status2 == TCL_ERROR && odbcOptLibNames[i] != NULL; ++i) {
	    path = Tcl_NewStringObj(odbcOptLibNames[i], -1);
	    Tcl_AppendObjToObj(path, shlibext);
	    Tcl_IncrRefCount(path);
	    status2 = Tcl_LoadFile(interp, path, NULL, 0, NULL, handle2Ptr);
	    if (status2 == TCL_OK) {
		SQLConfigDataSourceW =
		    (BOOL (INSTAPI*)(HWND, WORD, LPCWSTR, LPCWSTR))
		    Tcl_FindSymbol(NULL, *handle2Ptr, "SQLConfigDataSourceW");
		if (SQLConfigDataSourceW == NULL) {
		    SQLConfigDataSource =
			(BOOL (INSTAPI*)(HWND, WORD, LPCSTR, LPCSTR))
			Tcl_FindSymbol(NULL, *handle2Ptr,
				       "SQLConfigDataSource");
		}
		SQLInstallerError =
		    (BOOL (INSTAPI*)(WORD, DWORD*, LPSTR, WORD, WORD*))
		    Tcl_FindSymbol(NULL, *handle2Ptr, "SQLInstallerError");
	    } else {
		Tcl_ResetResult(interp);
	    }
	    Tcl_DecrRefCount(path);
	}
    }

    /*
     * Either we've successfully loaded a library (status == TCL_OK),
     * or we've run out of library names (in which case status==TCL_ERROR
     * and the error message reflects the last unsuccessful load attempt).
     */
    Tcl_DecrRefCount(shlibext);
    if (status != TCL_OK) {
	return NULL;
    }
    return handle;
}

#endif
