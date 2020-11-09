# Ansible Azure Deployment

Quickly deploy a demo environnement on Azure.

1. Generate an ssh key to be used with this demo environnement `ssh-keygen -t ed25519 -C "dns-tunnel-demo" -f creds/ssh/id_rsa`

2. Deploy the environnement in Azure with Ansible you can use the docker-compose.yml if Ansible has issue with the dependencies. You must be connected to azure with the `az login` before executing `ansible-playbook azure.yml`

3. Connect to the bastion with the script in `scripts/connect.sh`

3. Deploy the VMs `ansible-playbook config-all.yml -i hosts.yml`