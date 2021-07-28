#include <iostream>
#include <fstream>
#include <cmath>

using namespace std;

int main(void)
{
    float counter = 1000;

    ofstream myfile;
    myfile.open ("roots.txt");
        while (counter > 0)
        {
            myfile << counter << ", " << cos(counter) << endl;
            counter = counter - 1;
        }
    myfile.close();

    return 0;
}
