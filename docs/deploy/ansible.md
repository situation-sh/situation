______________________________________________________________________

## title: Ansible summary: Deploy situation on your infrastructure

Agents can be deployed automatically through [Ansible](https://docs.ansible.com/).
Given an [inventory](https://docs.ansible.com/ansible/latest/inventory_guide/index.html)
you can run the playbook given below.

```shell
ansible-playbook situation.yml -i inventory.yml
```

The playbook can we tuned to customize where the agent will be installed.
First it downloads locally the latest linux and windows binaries.
Then it deploys the binaries to the hosts defined in the inventory and outputs a list of agentds deployed.

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
        url: "https://api.github.com/repos/situation-sh/situation/releases/latest"
        return_content: true
        headers:
          Accept: "application/vnd.github+json"
      register: latest_release

    - name: Download matched release assets
      get_url:
        url: "{{ item.browser_download_url }}"
        dest: "{{ download_dir }}/{{ item.name }}"
        mode: "0644"
      loop: "{{ latest_release.json.assets }}"
      when: ("-linux" in item.name or "-windows.exe" in item.name) and
        (not lookup('file', download_dir + '/' + item.name, errors='ignore'))

- name: Deploy agents to hosts
  tags: deploy
  hosts: all
  gather_facts: false
  vars:
    src_dir: "{{ TMP_DIR | default('/tmp/situation') }}"
    api_key: "{{ SITUATION_API_KEY | default('') }}"
    dest_linux: "~/situation"
    dest_windows: "~/situation.exe"
    agent_ids: "{{ src_dir }}/agents.txt"
  tasks:
    - name: Filter and return only selected facts
      ansible.builtin.setup:
        filter:
          - "ansible_system"
          - "ansible_os_family"
          - "ansible_hostname"

    - set_fact:
        is_windows: >-
          {{ ansible_os_family == "Windows"  }}
        is_linux: >-
          {{ ansible_system == "Linux" }}

    - name: Copy Linux binary to target
      copy:
        src: "{{ src_dir }}/{{ item }}"
        dest: "{{ dest_linux }}"
        mode: "0755"
      loop: "{{ lookup('fileglob', src_dir + '/*-linux', wantlist=True) | map('basename') | list }}"
      when: is_linux

    - name: Copy Windows binary to target
      win_copy:
        src: "{{ src_dir }}/{{ item }}"
        dest: "{{ dest_windows }}"
      loop: "{{ lookup('fileglob', src_dir + '/*-windows.exe', wantlist=True) | map('basename') | list }}"
      when: is_windows

    - name: Define situation
      set_fact:
        situation: >-
          {{ dest_windows if is_windows else dest_linux }}

    - name: Refresh agent id
      win_command: "{{ situation }} refresh-id"
      when: is_windows

    - name: Refresh agent id
      command: "{{ situation }} refresh-id"
      when: is_linux

    - name: Print new id
      win_command: "{{ situation }} id"
      when: is_windows
      register: id_output_windows

    - name: Print new id
      command: "{{ situation }} id"
      when: is_linux
      register: id_output_linux

    - name: Retrieve new agent id
      set_fact:
        agent_id: >-
          {{ id_output_windows.stdout_lines[0] if is_windows else id_output_linux.stdout_lines[0] }}

    - name: List all deployments
      local_action:
        module: lineinfile
        line: "{{ ansible_hostname }},{{ agent_id }}"
        path: "{{ agent_ids }}"
        create: yes
```

{% endraw %}

## situation.sh

If you use the [situation.sh](https://situation.sh) platform you must authorize the deployed agents.
