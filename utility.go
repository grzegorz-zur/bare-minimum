package main

func wrap(position, length, step int, direction Direction) int {
	return ((position+step*int(direction))%length + length) % length
}
