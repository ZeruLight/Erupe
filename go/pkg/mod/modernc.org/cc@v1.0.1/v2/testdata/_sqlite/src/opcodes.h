/************** Begin file opcodes.h *****************************************/
/* Automatically generated.  Do not edit */
/* See the tool/mkopcodeh.tcl script for details */
#define OP_Savepoint       0
#define OP_AutoCommit      1
#define OP_Transaction     2
#define OP_SorterNext      3 /* jump                                       */
#define OP_PrevIfOpen      4 /* jump                                       */
#define OP_NextIfOpen      5 /* jump                                       */
#define OP_Prev            6 /* jump                                       */
#define OP_Next            7 /* jump                                       */
#define OP_Checkpoint      8
#define OP_JournalMode     9
#define OP_Vacuum         10
#define OP_VFilter        11 /* jump, synopsis: iplan=r[P3] zplan='P4'     */
#define OP_VUpdate        12 /* synopsis: data=r[P3@P2]                    */
#define OP_Goto           13 /* jump                                       */
#define OP_Gosub          14 /* jump                                       */
#define OP_InitCoroutine  15 /* jump                                       */
#define OP_Yield          16 /* jump                                       */
#define OP_MustBeInt      17 /* jump                                       */
#define OP_Jump           18 /* jump                                       */
#define OP_Not            19 /* same as TK_NOT, synopsis: r[P2]= !r[P1]    */
#define OP_Once           20 /* jump                                       */
#define OP_If             21 /* jump                                       */
#define OP_IfNot          22 /* jump                                       */
#define OP_IfNullRow      23 /* jump, synopsis: if P1.nullRow then r[P3]=NULL, goto P2 */
#define OP_SeekLT         24 /* jump, synopsis: key=r[P3@P4]               */
#define OP_SeekLE         25 /* jump, synopsis: key=r[P3@P4]               */
#define OP_SeekGE         26 /* jump, synopsis: key=r[P3@P4]               */
#define OP_SeekGT         27 /* jump, synopsis: key=r[P3@P4]               */
#define OP_NoConflict     28 /* jump, synopsis: key=r[P3@P4]               */
#define OP_NotFound       29 /* jump, synopsis: key=r[P3@P4]               */
#define OP_Found          30 /* jump, synopsis: key=r[P3@P4]               */
#define OP_SeekRowid      31 /* jump, synopsis: intkey=r[P3]               */
#define OP_NotExists      32 /* jump, synopsis: intkey=r[P3]               */
#define OP_Last           33 /* jump                                       */
#define OP_IfSmaller      34 /* jump                                       */
#define OP_SorterSort     35 /* jump                                       */
#define OP_Sort           36 /* jump                                       */
#define OP_Rewind         37 /* jump                                       */
#define OP_IdxLE          38 /* jump, synopsis: key=r[P3@P4]               */
#define OP_IdxGT          39 /* jump, synopsis: key=r[P3@P4]               */
#define OP_IdxLT          40 /* jump, synopsis: key=r[P3@P4]               */
#define OP_IdxGE          41 /* jump, synopsis: key=r[P3@P4]               */
#define OP_RowSetRead     42 /* jump, synopsis: r[P3]=rowset(P1)           */
#define OP_Or             43 /* same as TK_OR, synopsis: r[P3]=(r[P1] || r[P2]) */
#define OP_And            44 /* same as TK_AND, synopsis: r[P3]=(r[P1] && r[P2]) */
#define OP_RowSetTest     45 /* jump, synopsis: if r[P3] in rowset(P1) goto P2 */
#define OP_Program        46 /* jump                                       */
#define OP_FkIfZero       47 /* jump, synopsis: if fkctr[P1]==0 goto P2    */
#define OP_IfPos          48 /* jump, synopsis: if r[P1]>0 then r[P1]-=P3, goto P2 */
#define OP_IfNotZero      49 /* jump, synopsis: if r[P1]!=0 then r[P1]--, goto P2 */
#define OP_IsNull         50 /* jump, same as TK_ISNULL, synopsis: if r[P1]==NULL goto P2 */
#define OP_NotNull        51 /* jump, same as TK_NOTNULL, synopsis: if r[P1]!=NULL goto P2 */
#define OP_Ne             52 /* jump, same as TK_NE, synopsis: IF r[P3]!=r[P1] */
#define OP_Eq             53 /* jump, same as TK_EQ, synopsis: IF r[P3]==r[P1] */
#define OP_Gt             54 /* jump, same as TK_GT, synopsis: IF r[P3]>r[P1] */
#define OP_Le             55 /* jump, same as TK_LE, synopsis: IF r[P3]<=r[P1] */
#define OP_Lt             56 /* jump, same as TK_LT, synopsis: IF r[P3]<r[P1] */
#define OP_Ge             57 /* jump, same as TK_GE, synopsis: IF r[P3]>=r[P1] */
#define OP_ElseNotEq      58 /* jump, same as TK_ESCAPE                    */
#define OP_DecrJumpZero   59 /* jump, synopsis: if (--r[P1])==0 goto P2    */
#define OP_IncrVacuum     60 /* jump                                       */
#define OP_VNext          61 /* jump                                       */
#define OP_Init           62 /* jump, synopsis: Start at P2                */
#define OP_Return         63
#define OP_EndCoroutine   64
#define OP_HaltIfNull     65 /* synopsis: if r[P3]=null halt               */
#define OP_Halt           66
#define OP_Integer        67 /* synopsis: r[P2]=P1                         */
#define OP_Int64          68 /* synopsis: r[P2]=P4                         */
#define OP_String         69 /* synopsis: r[P2]='P4' (len=P1)              */
#define OP_Null           70 /* synopsis: r[P2..P3]=NULL                   */
#define OP_SoftNull       71 /* synopsis: r[P1]=NULL                       */
#define OP_Blob           72 /* synopsis: r[P2]=P4 (len=P1)                */
#define OP_Variable       73 /* synopsis: r[P2]=parameter(P1,P4)           */
#define OP_Move           74 /* synopsis: r[P2@P3]=r[P1@P3]                */
#define OP_Copy           75 /* synopsis: r[P2@P3+1]=r[P1@P3+1]            */
#define OP_SCopy          76 /* synopsis: r[P2]=r[P1]                      */
#define OP_IntCopy        77 /* synopsis: r[P2]=r[P1]                      */
#define OP_ResultRow      78 /* synopsis: output=r[P1@P2]                  */
#define OP_CollSeq        79
#define OP_AddImm         80 /* synopsis: r[P1]=r[P1]+P2                   */
#define OP_RealAffinity   81
#define OP_Cast           82 /* synopsis: affinity(r[P1])                  */
#define OP_Permutation    83
#define OP_BitAnd         84 /* same as TK_BITAND, synopsis: r[P3]=r[P1]&r[P2] */
#define OP_BitOr          85 /* same as TK_BITOR, synopsis: r[P3]=r[P1]|r[P2] */
#define OP_ShiftLeft      86 /* same as TK_LSHIFT, synopsis: r[P3]=r[P2]<<r[P1] */
#define OP_ShiftRight     87 /* same as TK_RSHIFT, synopsis: r[P3]=r[P2]>>r[P1] */
#define OP_Add            88 /* same as TK_PLUS, synopsis: r[P3]=r[P1]+r[P2] */
#define OP_Subtract       89 /* same as TK_MINUS, synopsis: r[P3]=r[P2]-r[P1] */
#define OP_Multiply       90 /* same as TK_STAR, synopsis: r[P3]=r[P1]*r[P2] */
#define OP_Divide         91 /* same as TK_SLASH, synopsis: r[P3]=r[P2]/r[P1] */
#define OP_Remainder      92 /* same as TK_REM, synopsis: r[P3]=r[P2]%r[P1] */
#define OP_Concat         93 /* same as TK_CONCAT, synopsis: r[P3]=r[P2]+r[P1] */
#define OP_Compare        94 /* synopsis: r[P1@P3] <-> r[P2@P3]            */
#define OP_BitNot         95 /* same as TK_BITNOT, synopsis: r[P1]= ~r[P1] */
#define OP_Column         96 /* synopsis: r[P3]=PX                         */
#define OP_String8        97 /* same as TK_STRING, synopsis: r[P2]='P4'    */
#define OP_Affinity       98 /* synopsis: affinity(r[P1@P2])               */
#define OP_MakeRecord     99 /* synopsis: r[P3]=mkrec(r[P1@P2])            */
#define OP_Count         100 /* synopsis: r[P2]=count()                    */
#define OP_ReadCookie    101
#define OP_SetCookie     102
#define OP_ReopenIdx     103 /* synopsis: root=P2 iDb=P3                   */
#define OP_OpenRead      104 /* synopsis: root=P2 iDb=P3                   */
#define OP_OpenWrite     105 /* synopsis: root=P2 iDb=P3                   */
#define OP_OpenDup       106
#define OP_OpenAutoindex 107 /* synopsis: nColumn=P2                       */
#define OP_OpenEphemeral 108 /* synopsis: nColumn=P2                       */
#define OP_SorterOpen    109
#define OP_SequenceTest  110 /* synopsis: if( cursor[P1].ctr++ ) pc = P2   */
#define OP_OpenPseudo    111 /* synopsis: P3 columns in r[P2]              */
#define OP_Close         112
#define OP_ColumnsUsed   113
#define OP_Sequence      114 /* synopsis: r[P2]=cursor[P1].ctr++           */
#define OP_NewRowid      115 /* synopsis: r[P2]=rowid                      */
#define OP_Insert        116 /* synopsis: intkey=r[P3] data=r[P2]          */
#define OP_InsertInt     117 /* synopsis: intkey=P3 data=r[P2]             */
#define OP_Delete        118
#define OP_ResetCount    119
#define OP_SorterCompare 120 /* synopsis: if key(P1)!=trim(r[P3],P4) goto P2 */
#define OP_SorterData    121 /* synopsis: r[P2]=data                       */
#define OP_RowData       122 /* synopsis: r[P2]=data                       */
#define OP_Rowid         123 /* synopsis: r[P2]=rowid                      */
#define OP_NullRow       124
#define OP_SeekEnd       125
#define OP_SorterInsert  126 /* synopsis: key=r[P2]                        */
#define OP_IdxInsert     127 /* synopsis: key=r[P2]                        */
#define OP_IdxDelete     128 /* synopsis: key=r[P2@P3]                     */
#define OP_DeferredSeek  129 /* synopsis: Move P3 to P1.rowid if needed    */
#define OP_IdxRowid      130 /* synopsis: r[P2]=rowid                      */
#define OP_Destroy       131
#define OP_Real          132 /* same as TK_FLOAT, synopsis: r[P2]=P4       */
#define OP_Clear         133
#define OP_ResetSorter   134
#define OP_CreateBtree   135 /* synopsis: r[P2]=root iDb=P1 flags=P3       */
#define OP_SqlExec       136
#define OP_ParseSchema   137
#define OP_LoadAnalysis  138
#define OP_DropTable     139
#define OP_DropIndex     140
#define OP_DropTrigger   141
#define OP_IntegrityCk   142
#define OP_RowSetAdd     143 /* synopsis: rowset(P1)=r[P2]                 */
#define OP_Param         144
#define OP_FkCounter     145 /* synopsis: fkctr[P1]+=P2                    */
#define OP_MemMax        146 /* synopsis: r[P1]=max(r[P1],r[P2])           */
#define OP_OffsetLimit   147 /* synopsis: if r[P1]>0 then r[P2]=r[P1]+max(0,r[P3]) else r[P2]=(-1) */
#define OP_AggStep0      148 /* synopsis: accum=r[P3] step(r[P2@P5])       */
#define OP_AggStep       149 /* synopsis: accum=r[P3] step(r[P2@P5])       */
#define OP_AggFinal      150 /* synopsis: accum=r[P1] N=P2                 */
#define OP_Expire        151
#define OP_TableLock     152 /* synopsis: iDb=P1 root=P2 write=P3          */
#define OP_VBegin        153
#define OP_VCreate       154
#define OP_VDestroy      155
#define OP_VOpen         156
#define OP_VColumn       157 /* synopsis: r[P3]=vcolumn(P2)                */
#define OP_VRename       158
#define OP_Pagecount     159
#define OP_MaxPgcnt      160
#define OP_PureFunc0     161
#define OP_Function0     162 /* synopsis: r[P3]=func(r[P2@P5])             */
#define OP_PureFunc      163
#define OP_Function      164 /* synopsis: r[P3]=func(r[P2@P5])             */
#define OP_CursorHint    165
#define OP_Noop          166
#define OP_Explain       167

/* Properties such as "out2" or "jump" that are specified in
** comments following the "case" for each opcode in the vdbe.c
** are encoded into bitvectors as follows:
*/
#define OPFLG_JUMP        0x01  /* jump:  P2 holds jmp target */
#define OPFLG_IN1         0x02  /* in1:   P1 is an input */
#define OPFLG_IN2         0x04  /* in2:   P2 is an input */
#define OPFLG_IN3         0x08  /* in3:   P3 is an input */
#define OPFLG_OUT2        0x10  /* out2:  P2 is an output */
#define OPFLG_OUT3        0x20  /* out3:  P3 is an output */
#define OPFLG_INITIALIZER {\
/*   0 */ 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01, 0x01,\
/*   8 */ 0x00, 0x10, 0x00, 0x01, 0x00, 0x01, 0x01, 0x01,\
/*  16 */ 0x03, 0x03, 0x01, 0x12, 0x01, 0x03, 0x03, 0x01,\
/*  24 */ 0x09, 0x09, 0x09, 0x09, 0x09, 0x09, 0x09, 0x09,\
/*  32 */ 0x09, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,\
/*  40 */ 0x01, 0x01, 0x23, 0x26, 0x26, 0x0b, 0x01, 0x01,\
/*  48 */ 0x03, 0x03, 0x03, 0x03, 0x0b, 0x0b, 0x0b, 0x0b,\
/*  56 */ 0x0b, 0x0b, 0x01, 0x03, 0x01, 0x01, 0x01, 0x02,\
/*  64 */ 0x02, 0x08, 0x00, 0x10, 0x10, 0x10, 0x10, 0x00,\
/*  72 */ 0x10, 0x10, 0x00, 0x00, 0x10, 0x10, 0x00, 0x00,\
/*  80 */ 0x02, 0x02, 0x02, 0x00, 0x26, 0x26, 0x26, 0x26,\
/*  88 */ 0x26, 0x26, 0x26, 0x26, 0x26, 0x26, 0x00, 0x12,\
/*  96 */ 0x00, 0x10, 0x00, 0x00, 0x10, 0x10, 0x00, 0x00,\
/* 104 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,\
/* 112 */ 0x00, 0x00, 0x10, 0x10, 0x00, 0x00, 0x00, 0x00,\
/* 120 */ 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x04, 0x04,\
/* 128 */ 0x00, 0x00, 0x10, 0x10, 0x10, 0x00, 0x00, 0x10,\
/* 136 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06,\
/* 144 */ 0x10, 0x00, 0x04, 0x1a, 0x00, 0x00, 0x00, 0x00,\
/* 152 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10,\
/* 160 */ 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,\
}

/* The sqlite3P2Values() routine is able to run faster if it knows
** the value of the largest JUMP opcode.  The smaller the maximum
** JUMP opcode the better, so the mkopcodeh.tcl script that
** generated this include file strives to group all JUMP opcodes
** together near the beginning of the list.
*/
#define SQLITE_MX_JUMP_OPCODE  62  /* Maximum JUMP opcode */

/************** End of opcodes.h *********************************************/
