#include <iostream>
using namespace std;
int main()
{
    int i, my_arr[10];

    for(i = 0; i < 10; i++)
    {
        cout << "number " << i << ": ";
        cin >> my_arr[i];
    }

    for(i = 0; i < 10; i++)
    {
        cout << my_arr[i] << " ";
    }

    for(i = 0; i < 10; i++)
    {
        cout << my_arr[i] / 2 << " ";
    }

    return 0;
}
