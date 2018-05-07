pragma solidity ^0.4.0;
contract HelloWorld{


 function getMaxNumber(uint a,uint b,uint c,uint d) constant returns (uint){

     var maxNumber = a;
     if(maxNumber < b)
     {
         maxNumber = b;
     }

     if(maxNumber < c)
     {
         maxNumber = c;
     }

     if(maxNumber < d)
     {
         maxNumber = d;
     }

     return maxNumber;


 }



}
