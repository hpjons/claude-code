#include <iostream>
using namespace std;

int main()
{
 float heildarsala, mestasala;
 int vikunumer;
 float sala[4];

     for(vikunumer = 0; vikunumer <= 5; vikunumer++)
     {
         cout << "Insert the sale of week nr. " << vikunumer << " : ";
         cin >> sala[vikunumer];
     }

    vikunumer = 1;
    heildarsala = 0;

    for(vikunumer = 0; vikunumer <= 5; vikunumer++)
    {
        heildarsala = heildarsala + sala[vikunumer];
    }

    // Finnum hæstu söluna

    mestasala = 0;
    for(vikunumer = 0; vikunumer <= 5; vikunumer++)
    {
        if (sala[vikunumer] >= mestasala)
        {
            mestasala = sala[vikunumer];
        }
    }

    cout << "Total sale amount of this month was : " << heildarsala << endl;
    cout << "The highest sale amount was : " << mestasala << endl;

 return 0;
}
