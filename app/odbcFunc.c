//
// Created by Abhijeet Padwal on 15/10/19.
//
#include <stdio.h>
#include <sql.h>
#include <sqlext.h>
#include <string.h>
#include <stdlib.h>

/*
 * to list installed drivers
 * refers to odbcinst.ini file
 * SQLDrivers(envHandle, direction, outDriverString, lenOutDriverString, outAttrString, lenOutAttrString)
*/

void listDrivers(SQLHENV env, SQLRETURN ret)
{
    char driver[256], attr[256];
    SQLSMALLINT driver_ret, attr_ret, direction;

    direction = SQL_FETCH_FIRST;
    printf("\n#######\tDrivers and attributes\t########");
    while (SQL_SUCCEEDED(ret = SQLDrivers(env, direction,
                                          driver, sizeof(driver), &driver_ret,
                                          attr, sizeof(attr), &attr_ret)))
    {
        direction = SQL_FETCH_NEXT;
        printf("\n%s - %s", driver, attr);
    }
    printf("\n");
}

/*
 * to list Data Source name
 * refers to odbc.ini file
 * SQLDataSources(envHandle, direction, outDSNString, lenOutDSNString, dsnRet, outDESCString, lenOutDESCString, descRet)
 */

void listDSN(SQLHENV env, SQLRETURN ret)
{
    char dsn[256], desc[256];
    SQLSMALLINT dsn_ret, desc_ret, direction;

    direction = SQL_FETCH_FIRST;
    printf("\n\n#########\tDSN and Description\t#########");
    while (SQL_SUCCEEDED(
        ret = SQLDataSources(env, direction, dsn, sizeof(dsn), &dsn_ret, desc, sizeof(desc), &desc_ret)))
    {
        direction = SQL_FETCH_NEXT;
        printf("\n%s - %s", dsn, desc);
    }
    printf("\n");
}

/*
 * to connect to as DATA SOURCE
 * can be used to connect to data sources listed in odbc.ini
 * SQLDriverConnect(DBHandle, windowHandle, dataSourceName, lenDataSouceName, outConnectionString,
 * 						sizeofOutConnectionString, lenOutString, driverComplete)
*/

void connectDSN(SQLHDBC dbc, char *dsname, SQLRETURN ret)
{
    SQLSMALLINT lenDbmsName, outstrlen;
    char outstr[1024];
    char dbms_name[1024];

    ret = SQLDriverConnect(dbc, NULL, dsname, SQL_NTS, (SQLPOINTER)outstr, sizeof(outstr), &outstrlen,
                           SQL_DRIVER_COMPLETE);
    if (SQL_SUCCEEDED(ret))
    {
        printf("\nConnected. \nConnection string is: \t%s", outstr);
        ret = SQLGetInfo(dbc, SQL_DATABASE_NAME, (SQLPOINTER)dbms_name, sizeof(dbms_name), &lenDbmsName);
        if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO)
        {
            printf("\nDatabase Name is: \t%s", dbms_name);
        }
    }

    printf("\n");
    //SQLDisconnect(dbc);
}

/*
 * to list the tables in given database
 * SQLTables(statementHandle, catalogName, lenCatalogName, schemaName, lenSchemaName, tableName, lenTableName)
*/

void listTables(SQLHDBC dbc, SQLRETURN ret)
{
    printf("\n#####list of available tables#####");
    SQLHSTMT stmt;
    char tabbuf[255];
    int i = 1;
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    SQLTables(stmt, NULL, 0, NULL, 0, NULL, 0, "TABLE", SQL_NTS);
    while (SQL_SUCCEEDED(ret = SQLFetch(stmt)))
    {
        SQLLEN indicator;
        ret = SQLGetData(stmt, 3, SQL_C_CHAR, tabbuf, sizeof(tabbuf), &indicator);
        if (SQL_SUCCEEDED(ret))
        {
            printf("\n%d: %s", i, tabbuf);
            i++;
        }
    }
    printf("\n");
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
}

/*
 * to list tables in given table
 * SQLColumns(statementHandle, catalogName, lenCatalogName, schemaName, lenSchemaName,
 * 					tableName, lenTableName, columnName, lenColumnName)
*/

void listColumns(SQLHDBC dbc, char *tableName, SQLRETURN ret)
{
    SQLHSTMT stmt;
    SQLLEN lenColumnName, lenTypeName, lenColumnSize, lenNullable;
    SQLCHAR strColumnName[128], strTypeName[128];
    SQLINTEGER ColumnSize;
    SQLSMALLINT Nullable;

    printf("\n\n######Description for table '%s'######", tableName);
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    ret = SQLColumns(stmt, NULL, 0, NULL, 0, (SQLCHAR *)tableName, SQL_NTS, NULL, 0);
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO)
    {
        SQLBindCol(stmt, 4, SQL_C_CHAR, strColumnName,
                   128, &lenColumnName);
        SQLBindCol(stmt, 6, SQL_C_CHAR, strTypeName,
                   128, &lenTypeName);
        SQLBindCol(stmt, 7, SQL_C_SLONG, &ColumnSize,
                   0, &lenColumnSize);
        SQLBindCol(stmt, 11, SQL_C_SSHORT, &Nullable,
                   0, &lenNullable);
        printf("\nColName\tColSize\tDtType\tNullable");
        while (SQL_SUCCEEDED(ret = SQLFetch(stmt)))
        {
            printf("\n%s\t%i\t%s\t%hd ", strColumnName, ColumnSize, strTypeName, Nullable);
        }
    }
    printf("\n");
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
}

/*
 * to list stored procedures
 * depending upon db implementation system, user defined procedures will be listed
 * SQLProcedures(statementHandle, catalogName, lenCatalogName, schemaName, lenSchemaName, procName, lenProcName)
 */

void listProcedures(SQLHDBC dbc, SQLRETURN ret)
{
    SQLHSTMT stmt;
    SQLCHAR strProcName[256];
    SQLLEN lenProcName;
    int i = 1;
    //    FILE *fptr;
    //    fptr = fopen("output.txt", "w");
    //    if (fptr == NULL){
    //        printf("fptr fails!!");
    //    }
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);

    ret = SQLProcedures(stmt, NULL, 0, NULL, 0, NULL, 0);
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO)
    {
        ret = SQLBindCol(stmt, 3, SQL_C_CHAR, strProcName, sizeof(strProcName), &lenProcName);
        printf("\nProcedures List: ");
        while (SQL_SUCCEEDED(ret = SQLFetch(stmt)))
        {
            printf("\n%d: %s", i, strProcName);
            i++;
            //            fprintf(fptr, "\n%s", strProcName);
        }
    }
    printf("\n");
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
    //    fclose(fptr);
}

/*
 * to list procedure definitions datatypes for system procedures/functions
 * SQLProcedureColumns(statementHandle, catalogName, lenCatalogName, schemaName, lenSchemaName, procName, lenProcName)
*/

void listProcedureColumns(SQLHDBC dbc, char *procName, SQLRETURN ret)
{
    SQLHSTMT stmt;
    SQLLEN lenProcedureName;
    SQLLEN lenColumnName;
    SQLLEN lenTypeName;
    SQLCHAR strProcedureName[256];
    SQLCHAR strColumnName[256];
    SQLCHAR strTypeName[256];

    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    ret = SQLProcedureColumns(stmt, NULL, 0, NULL, 0, procName, SQL_NTS, NULL, 0);
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO)
    {
        SQLBindCol(stmt, 3, SQL_C_CHAR, strProcedureName,
                   sizeof(strProcedureName), &lenProcedureName);
        SQLBindCol(stmt, 4, SQL_C_CHAR, strColumnName,
                   sizeof(strColumnName), &lenColumnName);
        SQLBindCol(stmt, 7, SQL_C_CHAR, strTypeName,
                   sizeof(strTypeName), &lenTypeName);
        printf("\nProcedure description...\nPrcName\t\tColName\t\tDtType");
        while (SQL_SUCCEEDED(ret = SQLFetch(stmt)))
        {
            printf("\n%s\t\t%s\t\t%s", strProcedureName, strColumnName, strTypeName);
        }
        printf("\n");
    }
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
}

void nativeSQLQuery(SQLHDBC dbc, char *query, SQLRETURN ret)
{
    printf("query is %s", query);
    SQLHSTMT stmt;
    SQLCHAR queryOut[1024];
    SQLINTEGER queryOutLen;
    SQLULEN attrnoscan;
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    ret = SQLGetStmtAttr(stmt, SQL_ATTR_NOSCAN, &attrnoscan, 0, 0);
    printf("%ld", attrnoscan);
    SQLSetStmtAttr(stmt, SQL_ATTR_NOSCAN, (SQLPOINTER)SQL_NOSCAN_ON, 0);
    ret = SQLGetStmtAttr(stmt, SQL_ATTR_NOSCAN, &attrnoscan, 0, 0);
    printf("%ld", attrnoscan);
    ret = SQLNativeSql(dbc, query, strlen(query), queryOut, 1024, &queryOutLen);
    if (ret == SQL_SUCCESS)
    {
        printf("native sql string is \n%s", queryOut);
    }
    printf("\n");
}
void errorQuery(SQLRETURN ret, SQLHSTMT stmt)
{
    int n = 1;
    SQLCHAR SqlState[6], Msg[SQL_MAX_MESSAGE_LENGTH];
    SQLINTEGER NativeError;
    SQLSMALLINT j, MsgLen;
    SQLRETURN ret1;
    if ((ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) || ret != SQL_INVALID_HANDLE)
    {
        while ((ret1 = SQLGetDiagRec(SQL_HANDLE_STMT, stmt, n, SqlState, &NativeError, Msg, sizeof(Msg), &MsgLen)) != SQL_NO_DATA)
        {
            printf("\nError is %s", Msg);
            n++;
        }
    }
}

void QueryError(SQLHDBC dbc, char *query, SQLRETURN ret)
{
    SQLHSTMT stmt;
    SQLRETURN ret1, ret2;

    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    ret1 = SQLExecDirect(stmt, query, strlen(query));
    printf("Query is %s", query);

    if (ret1 == SQL_ERROR)
    {
        errorQuery(ret1, stmt);
    }
}

void preparedExecQuery(SQLHDBC dbc, char *query, SQLRETURN ret)
{
    SQLHSTMT stmt;
    SQLSMALLINT numCols = 0, numAttrs = 0;

    printf("\nQuery is : %s", query);

    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    SQLPrepare(stmt, query, strlen(query));

    if (ret == SQL_SUCCESS)
    {
        ret = SQLNumResultCols(stmt, &numCols);
        if (ret == SQL_ERROR)
        {
            errorQuery(ret, stmt);
        }
        SQLNumParams(stmt, &numAttrs);
        SQLSMALLINT i;
        SQLCHAR *columnName[5];
        SQLSMALLINT columnNameLen[255];
        SQLSMALLINT columnDataType[255];
        SQLLEN columnDataSize[255];
        SQLSMALLINT columnDataDigits[255];
        SQLSMALLINT columnDataNullable[255];

        for (i = 0; i < numCols; i++)
        {
            columnName[i] = (SQLCHAR *)malloc(SQL_MAX_COLUMN_NAME_LEN);
            ret = SQLDescribeCol(stmt, i + 1, columnName[i], 255, &columnNameLen[i], &columnDataType[i],
                                 &columnDataSize[i], &columnDataDigits[i], &columnDataNullable[i]);
            if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO)
            {
                printf("\nColumn name : %s, Column Type: %d", columnName[i], columnDataType[i]);
            }
        }
        printf("\nNumber of columns present in result: %i", numCols);
        printf("\nNumber of parameters for query: %i\n", numAttrs);
    }
    printf("\n");
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
}

void directExecution(SQLHDBC dbc, char *query, SQLRETURN ret)
{
    SQLHSTMT stmt;
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    ret = SQLExecDirect(stmt, query, SQL_NTS);
    if (ret == SQL_ERROR)
    {
        errorQuery(ret, stmt);
    }
    SQLLEN cId = 0, lenPattern = 0, lenId = 0;
    SQLCHAR pattern[255];
    SQLBindCol(stmt, 1, SQL_C_SHORT, &cId, 2, (SQLPOINTER)&lenId);
    SQLBindCol(stmt, 2, SQL_C_CHAR, &pattern, 255, &lenPattern);
    while (SQL_SUCCEEDED(ret = SQLFetch(stmt)))
    {
        printf("\nId is %i Pattern is %s", (int)cId, pattern);
    }
    printf("\n");
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
}

void preparedExecInsert(SQLHDBC dbc, char *insert, SQLRETURN ret)
{
    SQLHSTMT stmt;
    SQLSMALLINT NumParams, DataType, DecimalDigits, Nullable;
    SQLULEN bytesRemaining;

    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    SQLSetStmtAttr(stmt, SQL_ATTR_CURSOR_TYPE, (SQLPOINTER)SQL_CURSOR_DYNAMIC, 0);
    ret = SQLPrepare(stmt, insert, SQL_NTS);
    if (ret == SQL_SUCCESS)
    {
        int i = 0;
        SQLNumParams(stmt, &NumParams);
        printf("\nNumber of params are %i", NumParams);
        for (i = 0; i < NumParams; i++)
        {
            ret = SQLDescribeParam(stmt, i + 1, &DataType, &bytesRemaining, &DecimalDigits, &Nullable);
            printf("\nData Type : %i, bytesRemaining : %i, DecimalDigits : %i, Nullable %i\n",
                   (int)DataType, (int)bytesRemaining,
                   (int)DecimalDigits, (int)Nullable);
        }
    }
    printf("\n");
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
}

int main()
{
    SQLHENV env;
    SQLHDBC dbc;
    SQLRETURN ret;
    int choice;
    char tableName[32], procName[32], statement[256], queryStatement[256], insertStatement[256];
    SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env);
    SQLSetEnvAttr(env, SQL_ATTR_ODBC_VERSION, (void *)SQL_OV_ODBC3, 0);
    SQLAllocHandle(SQL_HANDLE_DBC, env, &dbc);

    int quit = 0;
    printf("\n0. Exit\n1. List Drivers\n2. List Data Sources\n3. Connect DSN\n4. List Tables\n5. List Columns\n6. List Procedures"
           "\n7. List Procedure Columns\n8. Query Error\n9. Prepared Query\n10. Prepared Insert\n");
    while (!quit)
    {
        printf("\n Enter Choice: ");
        scanf("%i", &choice);
        switch (choice)
        {
        case 0:
            quit = 1;
            break;
        case 1:
            listDrivers(env, ret);
            break;
        case 2:
            listDSN(env, ret);
            break;
        case 3:
            connectDSN(dbc, "Driver={PostgreSQL UNICODE};Server=52.53.245.117;Port=5432;Database=university;Uid=admin;Pwd=Tibco123;", ret);
            break;
        case 4:
            listTables(dbc, ret);
            break;
        case 5:
            printf("\nTable Name: ");
            scanf("%s", tableName);
            listColumns(dbc, tableName, ret);
            break;
        case 6:
            listProcedures(dbc, ret);
            break;
        case 7:
            printf("\nProcedure Name: ");
            scanf("%s", procName);
            listProcedureColumns(dbc, procName, ret);
            break;
        case 8:
            printf("\nSQL Statement: ");
            scanf(" %[^\n]", statement);
            QueryError(dbc, statement, ret);
            break;
        case 9:
            printf("\nEnter SQL Query Statement: ");
            scanf(" %[^\n]", queryStatement);
            preparedExecQuery(dbc, queryStatement, ret);
            break;
        case 10:
            printf("\nSQL Insert Statement: ");
            scanf(" %[^\n]", insertStatement);
            preparedExecInsert(dbc, insertStatement, ret);
            break;
        default:
            printf("\ninvalid Choice");
            printf("\n0. Exit\n1. List Drivers\n2. List Data Sources\n3. Connect DSN\n4. List Tables\n5. List Columns\n6. List Procedures"
                   "\n7. List Procedure Columns\n8. Query Error\n9. Prepared Query\n10. Prepared Insert");
        }
    }

    SQLDisconnect(dbc);
    SQLFreeHandle(SQL_HANDLE_DBC, dbc);
    SQLFreeHandle(SQL_HANDLE_ENV, env);
    printf("\nDisconnected..");
    return 0;
}