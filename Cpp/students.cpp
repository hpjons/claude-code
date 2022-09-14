#include <iostream>
using namespace std;

int main() {
    const int MAX_STUDENTS=4;
    float studentGrades[ MAX_STUDENTS ] = { 1.10 };
    for (int i=0; i<MAX_STUDENTS; i++) {
        cout << i << " " << studentGrades[i] << '\n';
    }
    return 0;
}
