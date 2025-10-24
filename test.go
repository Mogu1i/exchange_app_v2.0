package main

import (
	"fmt"
)

// func twoSum(nums []int, target int) []int {
//     for i, x := range nums {
//         for j := i + 1; j < len(nums); j++ {
//             if x+nums[j] == target {
//                 return []int{i, j}
//             }
//         }
//     }
//     return nil
// }

func twoSum(nums []int, target int) []int {
	hashTable := map[int]int{}
	for i, x := range nums {
		if p, ok := hashTable[target-x]; ok {
			return []int{p, i}
		}
		hashTable[x] = i
	}
	return nil
}

func main() {
	a := []int{1, 2, 3, 4, 5, 6, 7}
	target := 9
	fmt.Println(twoSum(a, target))

}
