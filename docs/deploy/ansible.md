---
title: Ansible 
summary: Deploy situation on your infrastructure
---

Agents can be deployed automatically through [Ansible](https://docs.ansible.com/).
Given an [inventory](https://docs.ansible.com/ansible/latest/inventory_guide/index.html)
you can run the playbook given below.

```shell
ansible-playbook situation.yml -i inventory.yml
```

The playbook can be tuned to customize where the agent will be installed.
First it downloads locally the latest linux and windows binaries.
Then it deploys the binaries to the hosts defined in the inventory and outputs a list of agents deployed.

{% raw %}

```yaml
# situation.yml
- name: Download latest binaries (once)
  hosts: localhost
  gather_facts: false
  vars:
    repo: "situation-sh/situation"
    download_dir: "{{ TMP_DIR | default('/tmp/situation') }}"
  tasks:
    - name: Ensure download directory exists
      file:
        path: "{{ download_dir }}"
        state: directory

    - name: Get latest release from GitHub API
      uri:
        url: "https://api.github.com/repos/{{ repo }}/releases/latest"
        return_content: true
        headers:
          Accept: "application/vnd.github+json"
      register: latest_release

    - name: Download matched release assets
      get_url:
        url: "{{ item.browser_download_url }}"
        dest: "{{ download_dir }}/{{ item.name }}"
        mode: "0644"
        force: false
      loop: "{{ latest_release.json.assets }}"
      when: "'-linux' in item.name or '-windows.exe' in item.name"

- name: Deploy agents to hosts
  tags: deploy
  hosts: all
  gather_facts: false
  vars:
    src_dir: "{{ TMP_DIR | default('/tmp/situation') }}"
    dest_linux: "{{ ansible_env.HOME }}/situation"
    dest_windows: "{{ ansible_env.USERPROFILE }}\\situation.exe"
    agent_ids: "{{ src_dir }}/agents.txt"
    arch_map:
      x86_64: "amd64"
      aarch64: "arm64"
      armv5l: "armv5l"
      armv6l: "armv6l"
      armv7l: "armv7l"
      AMD64: "amd64"
  tasks:
    - name: Gather required facts
      ansible.builtin.setup:
        gather_subset:
          - "!all" # only the min subset is collected
          - "env"

    - name: Copy Linux binary to target
      copy:
        src: "{{ src_dir }}/{{ item }}"
        dest: "{{ dest_linux }}"
        mode: "0755"
      loop: "{{ lookup('fileglob', src_dir + '/*-' + arch_map[ansible_architecture] + '-linux', wantlist=True) | map('basename') | list }}"
      when: ansible_system == "Linux"

    - name: Copy Windows binary to target
      win_copy:
        src: "{{ src_dir }}/{{ item }}"
        dest: "{{ dest_windows }}"
      loop: "{{ lookup('fileglob', src_dir + '/*-' + arch_map[ansible_architecture] + '-windows.exe', wantlist=True) | map('basename') | list }}"
      when: ansible_os_family == "Windows"

    - name: Set situation binary path
      set_fact:
        situation: "{{ dest_windows if ansible_os_family == 'Windows' else dest_linux }}"

    - name: Refresh agent id (Windows)
      win_command: "{{ situation }} refresh-id"
      when: ansible_os_family == "Windows"

    - name: Refresh agent id (Linux)
      command: "{{ situation }} refresh-id"
      when: ansible_system == "Linux"

    - name: Get agent id (Windows)
      win_command: "{{ situation }} id"
      when: ansible_os_family == "Windows"
      register: id_output_windows
      changed_when: false

    - name: Get agent id (Linux)
      command: "{{ situation }} id"
      when: ansible_system == "Linux"
      register: id_output_linux
      changed_when: false

    - name: Retrieve new agent id
      set_fact:
        agent_id: "{{ (id_output_windows.stdout_lines[0]) if ansible_os_family == 'Windows' else (id_output_linux.stdout_lines[0]) }}"

    - name: List all deployments
      delegate_to: localhost
      throttle: 1
      lineinfile:
        line: "{{ ansible_hostname }},{{ agent_id }}"
        path: "{{ agent_ids }}"
        create: true
```

{% endraw %}

## situation.sh

If you use the [situation.sh](https://situation.sh) platform you must authorize the deployed agents.
