/** @file touchlog.h
 *  @brief Function prototypes for the touchlog tool.
 * 
 *  This contains function prototypes for the touchlog tool and all macros
 *  and constants needed.
 * 
 *  This file contains touchlog's main() function.
 * 
 *  @author Sasank 'squatch$' Vishnubhatla (sasank@vishnubhatlas.net)
 *  @bug No known bugs.
*/

#define FNAME_SIZE 15
#define DATE_MONTH_YEAR_SIZE 8

#define VERSION "0.0.2-alpha"
#define AUTHOR "Sasank 'squatch$' Vishnubhatla"
#define RELEASE_DATE "Tuesday, July 18, 2023"
#define HELP "touchlog\nA tool to make a logfile for a date\n\nOptions:\n\t-h\t\tDisplay this help message\n\t-d [mmddyyyy]\tMake a logfile for a specific date\n\t-v\t\tDisplay version information\n\t[noop]\t\tMake a logfile for the current date\n\nPlease report any bugs to Sasank Vishnubhatla at sasank@vishnubhatlas.net"

#define LOG_FMT "> month: %s\n> day: %s\n> year: %s\n\n|> events\n\n|> food\n\n|> emotions\n\n|> things to remember\n"

#define CUSTOM_REGEX_FMT "([0-9]{2})([0-9]{2})([0-9]{4})"
#define CUSTOM_REGEX_FMT_GROUPS (size_t)3

/** @brief Writes a logfile with file name based on the inputs
 * 
 *  Whenever handling if touchlog needs to write a file, the file name is
 *  based on a specific mmddyyyy format. Therefore, it is easy to abstract
 *  away the file writing from the input handling aspect of touchlog. So,
 *  this function handles writing a new log file with the file name based on
 *  the function parameters.
 * 
 *  @param day The day (dd) component of the mmddyyyy format
 *  @param month The month (mm) component of the mmddyyyy format
 *  @param year The year (yyyy) component of the mmddyyyy format
 *  @return Status code
*/
// [ ] TODO: update function signature to include optional path
int write_logfile(char day[3], char month[3], char year[5]);

/** @brief Handles the custom date input
 *
 *  Whenever touchlog is invoked, there is an opportunity for the user to
 *  require that a custom date in the mmddyyyy format is used instead of the
 *  current system time. When this option is invoked, the following
 *  method is invoked the handle parsing the raw input and verifying that it is
 *  in mmddyyyy format. This checking is done via a regular expression pattern.
 *
 *  @param raw The raw input from the console
 *  @return Status code
*/
int handle_custom(char *raw);

/** @brief Handles the normal date (non)input
 *
 *  Whenever touchlog is invoked, the normal usage is when the user does not
 *  input any additional options. Therefore, the normal case of using the
 *  current system date for the file name is used.
 *
 *  @param raw The raw input from the console
 *  @return Status code
*/
int handle_today();

/** @brief The main runner of touchlog
 *
 *  The main mechanism of touchlog.
 *
 *  @param argc The count of arguments
 *  @param argv The console inputted parameters stored as strings
 *  @return Status code
*/
int main(int argc, char *argv[]);
