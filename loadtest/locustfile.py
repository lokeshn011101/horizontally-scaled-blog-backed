import time
import random
from locust import HttpUser, task, between


class QuickstartUser(HttpUser):
    wait_time = between(1, 5)
    url = 'http://localhost:8082'

    @task(1)
    def get_profile(self):
        self.client.get(f"{self.url}/v1/users/{random.randint(1, 10)}")

    @task(2)
    def get_user_blogs(self):
        self.client.get(f"{self.url}/v1/users/{random.randint(1, 10)}/blogs")

    @task(3)
    def get_user_blogs(self):
        self.client.get(f"{self.url}/v1/blogs?userId={random.randint(1, 10)}")

    @task(6)
    def get_user_blogs(self):
        self.client.get(f"{self.url}/v1/blogs/{random.randint(1, 1000)}")
