# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"
  #config.vm.name = "mini-docker"

  config.vm.provider "virtualbox" do |v|
    v.name = "mini-docker"
  end

  # config.vm.synced_folder "../data", "/vagrant_data"

  config.vm.provider "virtualbox" do |vb|
  # Display the VirtualBox GUI when booting the machine
    vb.gui = false
    vb.memory = "1024"
  end


  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y golang docker.io curl wget git stress
    cp /vagrant/.bashrc /home/vagrant/.bashrc
  SHELL
end
