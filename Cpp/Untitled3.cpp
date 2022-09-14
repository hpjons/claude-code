#include <iostream>
using namespace std;

int main(void)
{
    float g;
    float t;
    float s;
    float s0;
    float v0;

    s0 = 0;
    v0 = 0;

    for (t = 0; t <= 20; t++)
    {
        g = 9.8;
        s = s0 + v0 * t + 0.5 * g * (t*t);
        cout << s << "   ";
    }

     for (t = 0; t <= 20; t++)
    {
        g = 1.6;
        s = s0 + v0 * t + 0.5 * g * (t*t);
        cout << s << " ";
    }


    return 0;
}
