package g42cloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/lbaas_v2/pools"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud"
)

func TestAccLBV2Member_basic(t *testing.T) {
	var member_1 pools.Member
	var member_2 pools.Member
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBV2MemberDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccLBV2MemberConfig_basic(rName),
				ExpectNonEmptyPlan: true, // Because admin_state_up remains false.
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBV2MemberExists("g42cloud_lb_member.member_1", &member_1),
					testAccCheckLBV2MemberExists("g42cloud_lb_member.member_2", &member_2),
				),
			},
			{
				Config:             testAccLBV2MemberConfig_update(rName),
				ExpectNonEmptyPlan: true, // Because admin_state_up remains false.
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("g42cloud_lb_member.member_1", "weight", "10"),
					resource.TestCheckResourceAttr("g42cloud_lb_member.member_2", "weight", "15"),
				),
			},
		},
	})
}

func testAccCheckLBV2MemberDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*huaweicloud.Config)
	elbClient, err := config.ElbV2Client(G42_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating G42Cloud elb client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "g42cloud_lb_member" {
			continue
		}

		poolId := rs.Primary.Attributes["pool_id"]
		_, err := pools.GetMember(elbClient, poolId, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Member still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckLBV2MemberExists(n string, member *pools.Member) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*huaweicloud.Config)
		elbClient, err := config.ElbV2Client(G42_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating G42Cloud elb client: %s", err)
		}

		poolId := rs.Primary.Attributes["pool_id"]
		found, err := pools.GetMember(elbClient, poolId, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Member not found")
		}

		*member = *found

		return nil
	}
}

func testAccLBV2MemberConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "g42cloud_vpc_subnet" "test" {
  name = "subnet-default"
}

resource "g42cloud_lb_loadbalancer" "loadbalancer_1" {
  name          = "%s"
  vip_subnet_id = data.g42cloud_vpc_subnet.test.subnet_id

}

resource "g42cloud_lb_listener" "listener_1" {
  name            = "%s"
  protocol        = "HTTP"
  protocol_port   = 8080
  loadbalancer_id = g42cloud_lb_loadbalancer.loadbalancer_1.id
}

resource "g42cloud_lb_pool" "pool_1" {
  name        = "%s"
  protocol    = "HTTP"
  lb_method   = "ROUND_ROBIN"
  listener_id = g42cloud_lb_listener.listener_1.id
}

resource "g42cloud_lb_member" "member_1" {
  address       = "192.168.0.10"
  protocol_port = 8080
  pool_id       = g42cloud_lb_pool.pool_1.id
  subnet_id     = data.g42cloud_vpc_subnet.test.subnet_id

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

resource "g42cloud_lb_member" "member_2" {
  address       = "192.168.0.11"
  protocol_port = 8080
  pool_id       = g42cloud_lb_pool.pool_1.id
  subnet_id     = data.g42cloud_vpc_subnet.test.subnet_id

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}
`, rName, rName, rName)
}

func testAccLBV2MemberConfig_update(rName string) string {
	return fmt.Sprintf(`
data "g42cloud_vpc_subnet" "test" {
  name = "subnet-default"
}

resource "g42cloud_lb_loadbalancer" "loadbalancer_1" {
  name          = "%s"
  vip_subnet_id = data.g42cloud_vpc_subnet.test.subnet_id
}

resource "g42cloud_lb_listener" "listener_1" {
  name            = "%s"
  protocol        = "HTTP"
  protocol_port   = 8080
  loadbalancer_id = g42cloud_lb_loadbalancer.loadbalancer_1.id
}

resource "g42cloud_lb_pool" "pool_1" {
  name        = "%s"
  protocol    = "HTTP"
  lb_method   = "ROUND_ROBIN"
  listener_id = g42cloud_lb_listener.listener_1.id
}

resource "g42cloud_lb_member" "member_1" {
  address        = "192.168.0.10"
  protocol_port  = 8080
  weight         = 10
  admin_state_up = "true"
  pool_id        = g42cloud_lb_pool.pool_1.id
  subnet_id      = data.g42cloud_vpc_subnet.test.subnet_id

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}

resource "g42cloud_lb_member" "member_2" {
  address        = "192.168.0.11"
  protocol_port  = 8080
  weight         = 15
  admin_state_up = "true"
  pool_id        = g42cloud_lb_pool.pool_1.id
  subnet_id      = data.g42cloud_vpc_subnet.test.subnet_id

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }
}
`, rName, rName, rName)
}
