# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  HOSTNAME = 'ceph-metric'

  config.vm.box = "ubuntu/trusty64"

  4.times.each do |i|
    config.vm.define "#{HOSTNAME}-#{i}" do |node|
      node.vm.define "#{HOSTNAME}-#{i}"
      node.vm.hostname = "#{HOSTNAME}-#{i}"

      node.vm.network :private_network, ip: "192.168.12.#{10+i}"

      node.vm.provider :virtualbox do |vb|
        2.times.each do |j|
          vb.customize [ "createhd", "--filename", "ceph-disk-#{i}-#{j}", "--size", "10000" ]
          vb.customize [ "storageattach", :id, "--storagectl", "SATAController", "--port", 3+j, "--device", 0, "--type", "hdd", "--medium", "ceph-disk-#{i}-#{j}.vdi" ]        
        end    
      end

      # node.vm.provision :ansible do |ansible|
      #   ansible.groups = {
      #     'mons' => [ 'ceph-metric-0', 'ceph-metric-1', 'ceph-metric-2' ],
      #     'osds' => [ 'ceph-metric-0', 'ceph-metric-1', 'ceph-metric-2', 'ceph-metric-3' ],
      #     'mdss' => [],
      #     'rgws' => [],
      #   }

      #   ansible.extra_vars = {
      #     'fsid'                    => '4a158d27-f750-41d5-9e7f-26ce4c9d2d45',
      #     'monitor_secret'          => 'AQAFx3RTAAAAABAAruXdSr8PTHAiTRgsyQMgPQ==',
      #     'cluster_network'         => '192.168.12.0/24',
      #     'public_network'          => '192.168.12.0/24',
      #     'monitor_interface'       => 'eth1',
      #     'devices'                 => ['/dev/sdb', '/dev/sdc'],
      #   }

      #   ansible.playbook = 'provision/provision.yml'
      # end
      
      # annoying vagrant-ansible provision. we just do this manually now.
    end
  end

end
