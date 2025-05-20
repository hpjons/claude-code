#include <iostream>
using namespace std;

int main()
{

int hours;
double gross_pay, rate, tax_rate;

cout << "Sláðu inn tímavinnukaup (ISK) : ";
cin >> rate;
cout << "Sláðu inn unnin tímafjölda (klst) : ";
cin >> hours;
cout << "Sláðu inn skattprósentu : ";
cin >> tax_rate;

if ( hours > 40 )
	gross_pay = rate * 40 + 1.5 * rate * ( hours - 40 );
else 
	gross_pay = rate * hours;

cout.setf(ios::fixed);
cout.setf(ios::showpoint);
cout.precision(2);

cout << "Klukkustundir = " << hours << endl;
cout << "Tímavinnukaup (ISK) =  " << rate << endl;
cout << "Heildarlaun fyrir skatt (ISK) = " << gross_pay << endl;
cout << "Heildarlaun eftir " << tax_rate << " skatt (ISK) = " << gross_pay - (gross_pay * tax_rate/100) << endl;
cout << "Frádráttur = " << gross_pay * (tax_rate/100) << endl; 
return 0;

}
