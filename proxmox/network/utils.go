package network

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
)

func UnpackId(id string) (string, string, error) {
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("invalid id %s", id)
	}
	return s[0], s[1], nil
}

func FormId(node string, name string) string {
	return fmt.Sprintf("%s/%s", node, name)
}

func GenerateInterfaceName(key string, existing []string) string {
	if len(existing) == 0 {
		return fmt.Sprintf("%s0", key)
	}

	var nums []int
	re := regexp.MustCompile(`\d+`)
	for _, e := range existing {
		num, _ := strconv.Atoi(re.FindString(e))
		nums = append(nums, num)
	}
	sort.Ints(nums)

	for i := 0; i < nums[len(nums)-1]+1; i++ {
		if !utils.Contains(nums, i) {
			return fmt.Sprintf("%s%d", key, i)
		}
	}
	n := nums[len(nums)-1] + 1
	return fmt.Sprintf("%s%d", key, n)
}
