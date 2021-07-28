#include<iostream>
#include<string>
#include<sstream>

using namespace std;
int main() {

 string input = "";

 // How to get a string/sentence with spaces
 cout << "Please enter a valid sentence (with spaces):\n>";
 getline(cin, input);
 cout << "You entered: " << input << endl << endl;

 return 0;
}
