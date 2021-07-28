#include <iostream>
#include <stdio.h>
using namespace std;
int main()
  {
	int number;

	cout << "Insert a number between 0 and 10 : ";
	cin >> number;

	while (number > 0)
    {
        printf("Down down down %i\n", number);
        number--;
    }

    while (number < 10)
    {
        cout << "Up up up " << number << endl;
        number++;
    }

	return 0;
  }
