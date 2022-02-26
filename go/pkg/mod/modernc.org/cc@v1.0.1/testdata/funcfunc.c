// Declare f as function (char) returning pointer to function (int) returning long.
//
// http://cdecl.ridiculousfish.com/?q=long+%28*f%28char%29%29+%28int%29%3B
//
long (*f(char c))(int i) { char d = c; }
