// Dags: 14.5.2019
// Höf: Helgi Páll Jónsson

#include<iostream>
using namespace std;
int main()
{
	float a, b;
	float test;

	cout << "Sláðu inn tölu 1 : ";
	cin >> a;
	cout << "Sláðu inn tölu 2 : ";
	cin >> b;
	cout << "tala 1 er " << a << endl;
	cout << "tala 2 er " << b << endl;
	
	test = a / b;

	if (test == 2){
		cout << a << " er helmingi stærri en " << b << endl;
	}
	else if (test == 0.5){
		cout << a << " er helmingi minni en " << b << endl;
	}
	else {
		cout << "Ekkert sérstaklega merkilegt við þessar tölur!" << endl;
	}

	cout << a << "/"<< b <<" = " << test << endl;


	return 0;
}

