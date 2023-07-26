/** @file ,touchlog.c
 *  @brief Implementation of touchlog
 *
 *  This file contains touchlog's main() function.
 *
 *  @author Sasank 'squatch$' Vishnubhatla (sasank@vishnubhatlas.net)
 *  @bug No known bugs.
*/

/* -- Includes -- */
#include ",touchlog.h"

#include <getopt.h>
#include <regex.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

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
int write_logfile(char day[3], char month[3], char year[5])
{
    char *fname = malloc(sizeof(char) * FNAME_SIZE);
    int fname_ret = sprintf(fname, "%s.%s.%s.log", month, day, year);

    if (!fname_ret)
    {
        return fname_ret;
    }

    FILE *nf = fopen(fname, "w+");
    if (nf == NULL)
    {
        // 134 = SIGABRT
        return 134;
    }

    char *logdata = malloc(sizeof(char) * (strlen(LOG_FMT) + DATE_MONTH_YEAR_SIZE));

    int logwrite_ret = fprintf(nf, LOG_FMT, month, day, year);
    if (!logwrite_ret)
    {
        return logwrite_ret;
    }

    printf("Wrote new logfile for today's date to %s\n", fname);

    return 0;
}

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
int handle_custom(char *raw)
{
    regex_t regex;
    regmatch_t groups[CUSTOM_REGEX_FMT_GROUPS];

    int c = regcomp(&regex, CUSTOM_REGEX_FMT, REG_EXTENDED);
    int s = regexec(&regex, raw, CUSTOM_REGEX_FMT_GROUPS, groups, 0);

    if (s != 0)
    {
        printf("%s\n", "Error: input is not in the format mmddyyyy");
        return s;
    }

    char day[3];
    char month[3];
    char year[5];

    for (unsigned int i = 0; i < CUSTOM_REGEX_FMT_GROUPS; i++)
    {
        if (groups[i].rm_so == (size_t)-1)
        {
            printf("%s\n", "Error: input is not in the format mmddyyyy");
            return 1;
        }
    }

    // [ ] TODO: make this smarter
    strncpy(month, raw, 2);
    strncpy(day, raw + 2, 2);
    strncpy(year, raw + 4, 4);

    day[2] = '\0';
    month[2] = '\0';
    year[4] = '\0';

    regfree(&regex);

    int write_ret = write_logfile(day, month, year);

    return write_ret;
}

/** @brief Handles the normal date (non)input
 *
 *  Whenever touchlog is invoked, the normal usage is when the user does not
 *  input any additional options. Therefore, the normal case of using the
 *  current system date for the file name is used.
 *
 *  @param raw The raw input from the console
 *  @return Status code
*/
int handle_today()
{
    time_t t = time(NULL);
    struct tm *tm = localtime(&t);

    char day[3];
    char month[3];
    char year[5];

    size_t day_ret = strftime(day, sizeof(day), "%d", tm);
    size_t month_ret = strftime(month, sizeof(month), "%m", tm);
    size_t year_ret = strftime(year, sizeof(year), "%Y", tm);

    if (!(day_ret && month_ret && year_ret))
    {
        return 1;
    }

    int write_ret = write_logfile(day, month, year);

    return write_ret;
}

/** @brief The main runner of touchlog
 *
 *  The main mechanism of touchlog.
 *
 *  @param argc The count of arguments
 *  @param argv The console inputted parameters stored as strings
 *  @return Status code
*/
int main(int argc, char *argv[])
{
    int opt;

    while ((opt = getopt(argc, argv, "hd:v")) != -1)
    {
        switch (opt)
        {
        case 'h':
            printf("%s\n", HELP);

            return 0;
        case 'd':
            int ret = handle_custom(optarg);

            if (ret != 0) {
                exit(ret);
            }

            return ret;
        case 'v':
            printf("%s\n", "touchlog");
            printf("Version: %s\n", VERSION);
            printf("Author : %s\n", AUTHOR);
            printf("Release date: %s\n", RELEASE_DATE);

            return 0;
        case '?':
            printf("%s\n", "Missing argument");

            return 0;
        }
    }

    int ret = handle_today();

    if (ret != 0) {
        exit(0);
    }

    return ret;
}
