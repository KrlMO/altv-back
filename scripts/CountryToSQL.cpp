#include <iostream>
#include <fstream>
#include <cstring>

using namespace std;

string escapeSQL(const char *s)
{
    string out;
    for (int i = 0; s[i]; i++)
    {
        if (s[i] == '\'')
            out += "''";
        else
            out += s[i];
    }
    return out;
}

int main()
{
    ifstream archPaises("paises.txt", ios::in);
    if (not archPaises.is_open())
    {
        cerr << "Nose puede abrir el archivo paises.txt" << endl;
        exit(1);
    }
    ofstream archSQL("paises.sql", ios::out);
    if (not archSQL.is_open())
    {
        cerr << "Nose puede abrir el archivo paises.sql" << endl;
        exit(1);
    }

    char cadena[100], c, *pais, *isocode;
    archSQL << "insert into al_country (name, iso_code)" << endl
            << "values ";
    int i = 0;
    while (1)
    {
        archPaises.getline(cadena, 100, ';');
        if (archPaises.eof())
            break;
        pais = new char[strlen(cadena) + 1];
        strcpy(pais, cadena);
        archPaises.getline(cadena, 20, ';');
        isocode = new char[strlen(cadena) + 1];
        strcpy(isocode, cadena);
        archPaises.getline(cadena, '\n');
        string paisEsc = escapeSQL(pais);
        string isoEsc = escapeSQL(isocode);

        archSQL << "('" << paisEsc << "', '" << isoEsc << "')," << endl;
    }
}