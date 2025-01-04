#include <iostream>
#include <unistd.h>

using namespace std;

int main()
{
    int start;

    do
    {
        cout << "Need Number to start the countdown (1 - 100): ";
        cin >> start;
    }
    while(!(start<100));

    do
    {
        cout << "T-minus %d\n" << start;
        start--;
        sleep(1);
    }
    while(start>0);

    cout << "Zero!\nBlast Off!\n";
    return 0;
}
