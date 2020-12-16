# MySQL in Brew

mysql: # Start Mysql
	$(title)
	@brew services start mysql

stop: # Stop Mysql
	$(title)
	@brew services stop mysql

shell: # Connect to Mysql
	$(title)
	@mysql -h$(MYSQL_HOSTNAME) -u$(MYSQL_USERNAME) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)

root: # Connect to Mysql as root
	$(title)
	@mysql -uroot

init: # Initialize Mysql database
	$(title)
	@mysql -uroot < sql/init.sql
