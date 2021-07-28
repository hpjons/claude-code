#include <iostream>
using namespace std;

int main()
{
    int number;

    do
    {
        cout << "Odd number? ";
        cin >> number;
            if (number % 2 == 0)
            {
                cout << "I said odd number dumbass!!\n";
            }
    } while (number % 2 == 0);

    {
        cout << "\nAH !! Finally!";
    }

    cout << "\nGreat, thanks, bye!";

    system("pause > 0");
}
